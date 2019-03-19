// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	master  bool
	version string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns proper master and node versions for the given inputs",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := container.NewClusterManagerClient(ctx)
		if err != nil {
			log.Fatalf("unable to create cluster manager client: %s", err)
		}

		if project == "" {
			log.Fatalf("Must specify a project")
		}

		if location == "" {
			log.Fatalf("Must specify a location")
		}

		if clusterName == "" {
			log.Fatalf("Must specify a cluster name")
		}

		cluster := cluster.NewManagedCluster(client, project, location, clusterName, 0)

		fmt.Printf("%v", cluster)

		if master {
			v, err := cluster.LatestMasterVersionForReleaseSeries(ctx, version)
			if err != nil {
				log.Fatalf("failed to get latest master versions: %s", err)
			}
			fmt.Fprintf(os.Stdout, v)
		}

		if cluster.Cluster == nil {
			fmt.Fprintf(os.Stderr, "cluster doesn't exist")
		}

		fmt.Fprintf(os.Stdout, cluster.Cluster.GetCurrentMasterVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	versionCmd.Flags().BoolVar(&master, "master", true, "Query for master version. Queries for node version if false")
	versionCmd.Flags().StringVar(&version, "version", "latest", "Query for master version. Queries for node version if false")
}
