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

package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/GoogleCloudPlatform/gke-rolling-updates-demo/placer/pkg/regions"
	"google.golang.org/api/compute/v1"

	"github.com/spf13/cobra"
)

var (
	machineType string
	project     string
	nodeCount   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "placer",
	Short: "Returns a GCP region with enough CPU quota to capacity requirements",
	Long:  `Returns a GCP region with enough CPU quota to capacity requirements`,
	Run: func(cmd *cobra.Command, args []string) {
		if machineType == "" {
			log.Fatalf("Please specify a machine type")
		}

		if project == "" {
			log.Fatalf("Please specify a project")
		}

		if nodeCount == -1 {
			log.Fatalf("Please specify a node count")
		}

		service, err := compute.NewService(context.Background())
		if err != nil {
			log.WithField("error", err).Fatalf("Unable to create compute service")
		}

		regions, err := regions.New(service, project)
		if err != nil {
			log.WithField("error", err).Fatalf("Unable to get regions")
		}

		regions.Regions = shuffle(regions.Regions)

		for _, region := range regions.Regions {
			clusterCpus, err := getTotalCpus(machineType, nodeCount)
			if err != nil {
				log.WithField("error", err).Fatalf("Unable to calculate total CPUs needed")
			}

			log.WithFields(log.Fields{
				"region":         region.Name,
				"available_cpus": (region.CPULimit() - region.CPUUsage()),
				"requested_cpus": clusterCpus,
			}).Info("Comparing CPU quota to request")

			if clusterCpus <= (region.CPULimit() - region.CPUUsage()) {
				var zones []string
				for _, zone := range region.Zones {
					fmtZone := strings.Split(zone, "/")
					zones = append(zones, fmtZone[len(fmtZone)-1])
				}
				_, err := fmt.Fprintf(os.Stdout, "%s;%s", region.Name, strings.Join(zones, ","))
				if err != nil {
					log.WithField("error", err).Fatalf("Unable to write to stdout")
				}
				break
			}

			log.WithField("region", region.Name).Info("Region does not have quota to meet request")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&machineType, "machine-type", "", "machine type to use to calculate required capacity")
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "project to calculate required capacity for")
	rootCmd.PersistentFlags().IntVar(&nodeCount, "node-count", -1, "number of nodes to use to calculate required capacity")
}

// getTotalCpus calculates the total number of CPUs used by a certain number of
// machines of a given machine type.
func getTotalCpus(machineType string, nodeCount int) (int, error) {
	cpus, err := cpusForMachineType(machineType)
	if err != nil {
		return cpus, fmt.Errorf("unable to get CPUs for machine type: %s", err)
	}
	return cpus * nodeCount, nil
}

// cpusForMachineType parses a machine type and returns the number of CPUs it
// uses.
func cpusForMachineType(machineType string) (int, error) {
	parsedMachineType := strings.Split(machineType, "-")

	if parsedMachineType[0] == "custom" {
		cpus, err := strconv.Atoi(parsedMachineType[1])
		if err != nil {
			return -1, fmt.Errorf("Unable to convert CPU count to an integer")
		}
		return cpus, nil
	}

	cpus, err := strconv.Atoi(parsedMachineType[2])
	if err != nil {
		return -1, fmt.Errorf("Unable to convert CPU count to an integer")
	}
	return cpus, nil

}

// shuffle randomizes a list of regions. It implements a Fisher-Yates shuffle
// using pseudo-random values generated generated from a seed based on the
// timestamp.
func shuffle(regions []*regions.Region) []*regions.Region {
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range regions {
		if i == 0 {
			continue
		}
		temp := regions[i]
		regions[i] = regions[rand.Intn(i)]
		regions[rand.Intn(i)] = temp
	}
	return regions
}
