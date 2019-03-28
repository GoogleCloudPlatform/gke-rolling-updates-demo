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
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

var (
	nodeCount int
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		if nodeCount <= 0 {
			log.Fatalf("Must specify node count greater than 0")
		}

		c, err := cluster.Get(client, project, location, clusterName)
		if c != nil {
			log.WithFields(log.Fields{
				"project":        project,
				"location":       location,
				"cluster_name":   clusterName,
				"status":         c.Status,
				"status_message": c.StatusMessage,
			}).Fatalf("cluster already exists")
		}

		gc, err := cluster.NewGKECluster(client, project, location, clusterName, int32(nodeCount))
		if err != nil {
			switch e := err.(type) {
			case cluster.Error:
				if e.ClusterStatus() == containerpb.Cluster_ERROR {
					log.Error("cluster in error state")
					os.Exit(2)
				}
				log.Fatalf("cluster in bad state: %s", e.ClusterStatus())
			default:
				log.Fatalf("cluster in bad state: %s", e.Error())
			}
		}

		_, err = fmt.Fprintf(os.Stdout, fmt.Sprintf("%s", gc.Cluster.GetCurrentMasterVersion()))
		if err != nil {
			log.Fatalf("Failed writing to stdout: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().IntVar(&nodeCount, "node-count", 0, "A help for foo")
}
