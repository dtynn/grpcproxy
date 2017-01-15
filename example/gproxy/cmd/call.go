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

	"github.com/dtynn/grpcproxy/example/gproxy/client"
)

// callCmd represents the call command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "call",
	Long:  `call`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		if len(args) < 3 {
			log.Fatalf("3 args required, got %d", len(args))
		}

		log.Println("calling", args)
		if err := client.Call(args[1], args[0], args[2]); err != nil {
			log.Fatal("got error", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(callCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// callCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// callCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
