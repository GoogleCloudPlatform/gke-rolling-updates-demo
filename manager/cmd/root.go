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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	project     string
	location    string
	clusterName string
)

var rootCmd = &cobra.Command{
	Use:   "manager",
	Short: "Provides an interface to more easily manage GKE clusters",
	Long:  ``,
}

// Execute is the entrypoint for the application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "GCP project to run against")
	rootCmd.PersistentFlags().StringVar(&location, "location", "", "region/zone to use")
	rootCmd.PersistentFlags().StringVar(&clusterName, "cluster-name", "", "name of cluster")
}
