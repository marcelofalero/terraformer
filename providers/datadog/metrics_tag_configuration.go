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
	datadogClientV2 := g.Args["clientV2"].(*datadogV2.APIClient)
	auth := g.Args["auth"].(context.Context)

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
		metricTagConfiguration, r, err := datadogClientV2.MetricsApi.GetMetricTagConfiguration(auth, metricName)
		if err != nil {
			if r != nil && r.StatusCode == 404 {
				log.Printf("Metric tag configuration for metric %s not found, skipping", metricName)
				continue
			}
			log.Printf("Failed to get metric tag configuration for %s: %v", metricName, err)
			return err
		}

		resource := g.createResource(metricName)
		attributes := metricTagConfiguration.GetData().GetAttributes()
		resource.AdditionalAttributes["metric_type"] = attributes.GetMetricType().String()
		if attributes.HasTags() {
			resource.AdditionalAttributes["tags"] = attributes.GetTags()
		}
		if attributes.HasIncludePercentiles() {
			resource.AdditionalAttributes["include_percentiles"] = attributes.GetIncludePercentiles()
		}

		g.Resources = append(g.Resources, resource)
	}

	return nil
}