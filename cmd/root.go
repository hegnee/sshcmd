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
	"fmt"
	"github.com/fanux/sealos/install"
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

var command, localFilePath, remoteFilePath, mode string
var host []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shell",
	Short: "A brief description of your application",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		validate()
		var wg sync.WaitGroup
		for _, node := range host {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				switch mode {
				case "ssh":
					install.Cmd(host, command)
				case "scp":
					install.Copy(host, localFilePath, remoteFilePath)
				case "ssh|scp":
					install.Cmd(host, command)
					install.Copy(host, localFilePath, remoteFilePath)
				case "scp|ssh":
					install.Copy(host, localFilePath, remoteFilePath)
					install.Cmd(host, command)
				default:
					install.Cmd(host, command)
				}

			}(node)
		}
		wg.Wait()
	},
}

//validate host is connect
func validate() {
	if len(host) == 0 {
		logger.Error("hosts not allow empty")
		os.Exit(1)
	}
	if install.User == "" {
		logger.Error("user not allow empty")
		os.Exit(1)
	}
	var session *ssh.Session
	var errors []error
	for _, h := range host {
		session, err := install.Connect(install.User, install.Passwd, install.PrivateKeyFile, h)
		if err != nil {
			logger.Error("[%s] ------------ check error", h)
			logger.Error("[%s] ------------ error[%s]", h, err)
			errors = append(errors, err)
		} else {
			logger.Info("[%s]  ------------ check ok", h)
			logger.Info("[%s]  ------------ session[%p]", h, session)
		}
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	if len(errors) > 0 {
		logger.Error("has some linux server is connection ssh is failed")
		os.Exit(1)
	}
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
	// Here you will define your flags and configuration settings.
	rootCmd.Flags().StringVar(&install.User, "user", "root", "servers user name for ssh")
	rootCmd.Flags().StringVar(&install.Passwd, "passwd", "", "password for ssh")
	rootCmd.Flags().StringVar(&install.PrivateKeyFile, "pk", "/root/.ssh/id_rsa", "private key for ssh")
	rootCmd.Flags().StringSliceVar(&host, "host", []string{}, "exec host")
	rootCmd.Flags().StringVar(&command, "cmd", "", "exec shell")
	rootCmd.Flags().StringVar(&localFilePath, "local-path", "", "local path , ex /etc/local.txt")
	rootCmd.Flags().StringVar(&remoteFilePath, "remote-path", "", "local path , ex /etc/local.txt")
	rootCmd.Flags().StringVar(&mode, "mode", "ssh", "mode type ,use | spilt . ex ssh scp ssh|scp scp|ssh")
}
