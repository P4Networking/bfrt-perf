// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

package p4rt

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
		conn, err = grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		grpcClients[host] = conn
		go MonitorConnection(conn)
	}
	return
}
