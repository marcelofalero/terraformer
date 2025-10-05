// Copyright 2023 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datadog

import (
	"context"
	"fmt"
	"log"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MetricsTagConfigurationGenerator ...
type MetricsTagConfigurationGenerator struct {
	DatadogService
}

func (g *MetricsTagConfigurationGenerator) createResource(metricName string) terraformutils.Resource {
	return terraformutils.NewResource(
		metricName,
		fmt.Sprintf("metrics_tag_configuration_%s", metricName),
		"datadog_metric_tag_configuration",
		"datadog",
		map[string]string{
			"metric_name": metricName,
		},
		[]string{},
		map[string]interface{}{},
	)
}

// InitResources Generate TerraformResources from Datadog API,
// from each metric create 1 TerraformResource.
// Need Metric Name as ID for terraform resource
func (g *MetricsTagConfigurationGenerator) InitResources() error {
	datadogClient := g.Args["datadogClient"].(*datadog.APIClient)
	auth := g.Args["auth"].(context.Context)
	api := datadogV2.NewMetricsApi(datadogClient)

	var metricNames []string
	for _, filter := range g.Filter {
		if filter.FieldPath == "id" && filter.IsApplicable("metrics_tag_configuration") {
			metricNames = append(metricNames, filter.AcceptableValues...)
		}
	}

	if len(metricNames) == 0 {
		log.Print("Filter(metric names as IDs) is required for importing datadog_metrics_tag_configuration resource")
		return nil
	}

	for _, metricName := range metricNames {
		metricTagConfiguration, r, err := api.ListTagConfigurationByName(auth, metricName)
		if err != nil {
			if r != nil && r.StatusCode == 404 {
				log.Printf("Metric tag configuration for metric %s not found, skipping", metricName)
				continue
			}
			log.Printf("Failed to get metric tag configuration for %s: %v", metricName, err)
			return err
		}

		resource := g.createResource(metricName)
		if data := metricTagConfiguration.Data; data != nil {
			if attributes := data.Attributes; attributes != nil {
				if attributes.MetricType != nil {
					resource.AdditionalFields["metric_type"] = string(*attributes.MetricType)
				}
				if len(attributes.Tags) > 0 {
					resource.AdditionalFields["tags"] = attributes.Tags
				}
				if attributes.IncludePercentiles != nil {
					resource.AdditionalFields["include_percentiles"] = *attributes.IncludePercentiles
				}
			}
		}

		g.Resources = append(g.Resources, resource)
	}

	return nil
}