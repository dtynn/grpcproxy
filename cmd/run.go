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

	"github.com/dtynn/grpcproxy/config"
	"github.com/dtynn/grpcproxy/service"
	"github.com/dtynn/grpcproxy/version"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run grpcproxy server",
	Long:  `run grpcproxy server with the given config file, default "./example.conf"`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("[GRPCPROXY] version %s", version.Version())

		if cfgFile == "" {
			cfgFile = "./example.conf"
		}

		cfg, err := config.ReadConfig(cfgFile)
		if err != nil {
			log.Fatalf("fail to read config: %s", err)
		}

		svr, err := service.NewService(cfg)
		if err != nil {
			log.Fatalf("fail to init service %s", err)
		}

		if err := svr.Run(); err != nil {
			log.Fatalf("server failure %s", err)
		}

		log.Println("server stopped")
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
