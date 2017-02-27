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
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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

		svr, err := service.NewServiceWithCfgFile(cfgFile)
		if err != nil {
			log.Fatalf("fail to init service %s", err)
		}

		go signalHandler(svr)

		if err := svr.Run(); err != nil {
			log.Fatalf("got server error %q", err)
		}

		log.Println("server stopped")
	},
}

func signalHandler(service *service.Service) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		log.Printf("[SERVER] got signal %s", sig)

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// this ensures a subsequent INT/TERM will trigger standard go behaviour of
			// terminating.
			signal.Stop(ch)
			service.Close()
			return

		case syscall.SIGUSR2:
			service.ReloadConfigFile()
		}
	}
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
