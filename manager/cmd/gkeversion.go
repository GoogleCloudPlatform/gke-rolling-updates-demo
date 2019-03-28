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
	"os"

	container "cloud.google.com/go/container/apiv1"
	"github.com/GoogleCloudPlatform/gke-rolling-updates-demo/manager/pkg/cluster"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	master     bool
	gkeVersion string
)

var gkeVersionCmd = &cobra.Command{
	Use:   "gke-version",
	Short: "Returns proper GKE master and node versions for the given inputs",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := container.NewClusterManagerClient(ctx)
		if err != nil {
			log.Fatalf("unable to create cluster manager client: %s", err)
		}

		cluster, err := cluster.NewGKECluster(client, project, location, clusterName, 0)
		if err != nil {
			log.Fatalf("error getting GKE cluster handle: %s", err)
		}

		if cluster.Cluster == nil {
			fmt.Fprintf(os.Stderr, "cluster doesn't exist\n")
		}

		_, err = fmt.Fprintf(os.Stdout, fmt.Sprintf("%s", cluster.Cluster.GetCurrentMasterVersion()))
		if err != nil {
			log.Fatalf("Failed writing to stdout: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(gkeVersionCmd)
	gkeVersionCmd.Flags().BoolVar(&master, "master", true, "Query for master version. Queries for node version if false")
	gkeVersionCmd.Flags().StringVar(&gkeVersion, "version", "latest", "Query for master version. Queries for node version if false")
}
