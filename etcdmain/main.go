// Copyright 2015 The etcd Authors
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

package etcdmain

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strings"
)

const (
	grpcCmd = "grpc-proxy"
	gatewayCmd = "gateway"
)

func (c *etcdCommand) GetChannels() (<- chan struct{}, <- chan error){
	return c.readyc, c.errorc
}

func (c *etcdCommand) Init() (*zap.Logger, *config, error) {
	return c.zl, c.cfg, c.err
}

func (c *etcdCommand) Start() error {
	return c.Execute()
}

func (c *etcdCommand) Stop() error {
	// todo: do we need to close/stop anything here?
	//       what about sockets/grpc client & server
	return nil
}

func Main() {
	checkSupportArch()

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if covArgs := os.Getenv("ETCDCOV_ARGS"); len(covArgs) > 0 {
			args := strings.Split(os.Getenv("ETCDCOV_ARGS"), "\xe7\xcd")[1:]
			rootCmd.SetArgs(args)
			cmd = grpcCmd
		}

		a := &abstractAdapter{}

		switch cmd {
		case grpcCmd:
			a.cmd = grpcProxyCmd
			break
		case gatewayCmd:
			a.cmd = gatewayCommand
			break
		}

		switch cmd {
		case grpcCmd, gatewayCmd:
			if err := Run(a); err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		}
	}

	_ = Run(&abstractAdapter{svc: svcAdapter})
}

