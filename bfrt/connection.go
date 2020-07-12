// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package bfrt

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Cache of address to gRPC client
var grpcClients = make(map[string]*grpc.ClientConn)

func MonitorConnection(conn *grpc.ClientConn) {
	state := conn.GetState()
	for {
		fmt.Printf("gRPC state update for %s: %v\n", conn.Target(), state.String())
		if state == connectivity.Shutdown {
			break
		}
		conn.WaitForStateChange(context.Background(), state)
		state = conn.GetState()
	}
}

func GetConnection(host string) (conn *grpc.ClientConn, err error) {
	conn, ok := grpcClients[host]
	if !ok {
		conn, err = grpc.Dial(host, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(128*1024*1024),
				grpc.MaxCallSendMsgSize(128*1024*1024)))
		if err != nil {
			return nil, err
		}
		grpcClients[host] = conn
		go MonitorConnection(conn)
	}
	return
}
