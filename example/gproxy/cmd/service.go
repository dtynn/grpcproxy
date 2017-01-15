// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"log"

	"github.com/spf13/cobra"

	"github.com/dtynn/grpcproxy/example/gproxy/bar"
	"github.com/dtynn/grpcproxy/example/gproxy/foo"
)

var port int

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "start service",
	Long:  `start a test service`,
	Run: func(cmd *cobra.Command, args []string) {
		if port <= 0 {
			log.Fatalf("port required")
		}

		// TODO: Work your own magic here
		if len(args) == 0 {
			log.Fatalf("service name required")
		}

		name := args[0]
		switch name {
		case "foo":
			s := foo.Service{
				Port: port,
			}

			if err := s.Run(); err != nil {
				log.Fatalf("foo service error: %s", err)
			}

		case "bar":
			s := bar.Service{
				Port: port,
			}

			if err := s.Run(); err != nil {
				log.Fatalf("bar service error: %s", err)
			}

		default:
			log.Fatalf("unsupport service name %s", name)
		}
	},
}

func init() {
	serviceCmd.Flags().IntVarP(&port, "port", "p", 0, "port")
	RootCmd.AddCommand(serviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
