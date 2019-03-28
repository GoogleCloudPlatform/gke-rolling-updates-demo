/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package regions

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
)

type Regions struct {
	Service *compute.Service
	Regions []*Region
}

func New(service *compute.Service, project string) (*Regions, error) {
	regionsListCall := service.Regions.List(project)

	var regions []*Region
	for {
		regionsListResponse, err := regionsListCall.Do()
		if err != nil {
			return nil, fmt.Errorf("Unable to get list of regions: %s", err)
		}

		for _, region := range regionsListResponse.Items {
			r := &Region{Region: region}
			log.WithFields(log.Fields{
				"name":      region.Name,
				"project":   project,
				"cpu_usage": r.CPUUsage(),
				"cpu_limit": r.CPULimit(),
			}).Info("Found region")
			regions = append(regions, r)
		}

		if regionsListResponse.NextPageToken == "" {
			break
		}
		regionsListCall = regionsListCall.PageToken(regionsListResponse.NextPageToken)
	}

	return &Regions{
		Service: service,
		Regions: regions,
	}, nil
}

type Region struct {
	*compute.Region
}

func (r *Region) CPUUsage() int {
	for _, v := range r.Quotas {
		if v.Metric == "CPUS" {
			return int(math.Round(v.Usage))
		}
	}
	return -1
}

func (r *Region) CPULimit() int {
	for _, v := range r.Quotas {
		if v.Metric == "CPUS" {
			return int(math.Round(v.Limit))
		}
	}
	return -1
}
