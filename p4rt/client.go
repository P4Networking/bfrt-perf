// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"context"
	"fmt"

	p4 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/genproto/googleapis/rpc/code"
)

var p4rtClients = make(map[p4rtClientKey]P4RuntimeClient)

type P4RuntimeClient interface {
	SetMastership(electionID p4.Uint128) error
	GetForwardingPipelineConfig() (*p4.ForwardingPipelineConfig, error)
	SetForwardingPipelineConfig(p4InfoPath, deviceConfigPath string) error
	Write(req *p4.WriteRequest) <-chan []*p4.Error
	SetWriteTraceChan(traceChan chan WriteTrace)
	DeviceID() uint64
	ElectionID() *p4.Uint128
}

type p4rtClientKey struct {
	host     string
	deviceID uint64
}

type p4rtClient struct {
	client         p4.P4RuntimeClient
	stream         p4.P4Runtime_StreamChannelClient
	deviceID       uint64
	electionID     p4.Uint128
	writes         chan p4Write
	writeTraceChan chan WriteTrace
	batchSize      int
	numThreads     int
}

func (c *p4rtClient) Init() (err error) {
	// Initialize stream for mastership and packet I/O
	c.stream, err = c.client.StreamChannel(context.Background())
	if err != nil {
		return
	}
	go func() {
		for {
			res, err := c.stream.Recv()
			if err != nil {
				fmt.Printf("stream recv error: %v\n", err)
			} else if arb := res.GetArbitration(); arb != nil {
				if code.Code(arb.Status.Code) == code.Code_OK {
					fmt.Println("client is master")
				} else {
					fmt.Println("client is not master")
				}
			} else {
				fmt.Printf("stream recv: %v\n", res)
			}

		}
	}()

	var writeBufferSize = c.batchSize * c.numThreads * 10
	// Initialize Write thread
	c.writes = make(chan p4Write, writeBufferSize)
	for i := 0; i < c.numThreads; i++ {
		go c.ListenForWrites()
	}

	return
}

func (c *p4rtClient) DeviceID() uint64 {
	return c.deviceID
}

func (c *p4rtClient) ElectionID() *p4.Uint128 {
	return &c.electionID
}

func CreateOrGetP4RuntimeClient(host string, deviceID uint64, batchSize int, numThreads int) (P4RuntimeClient, error) {
	key := p4rtClientKey{
		host:     host,
		deviceID: deviceID,
	}

	// First, return a P4RT client if one exists
	if p4rtClient, ok := p4rtClients[key]; ok {
		return p4rtClient, nil
	}

	// Second, check to see if we can reuse the gRPC connection for a new P4RT client
	conn, err := GetConnection(host)
	if err != nil {
		return nil, err
	}
	client := &p4rtClient{
		client:     p4.NewP4RuntimeClient(conn),
		deviceID:   deviceID,
		batchSize:  batchSize,
		numThreads: numThreads,
	}
	err = client.Init()
	if err != nil {
		return nil, err
	}
	p4rtClients[key] = client
	return client, nil
}
