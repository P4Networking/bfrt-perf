// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package bfrt

import (
	"context"
	"fmt"
	p4 "github.com/breezestars/go-bfrt/proto/out"
)

var bfrtClients = make(map[bfrtClientKey]BFRuntimeClient)

type BFRuntimeClient interface {
	SetMastership(clientId uint32) error
	GetForwardingPipelineConfig() ([]*p4.ForwardingPipelineConfig, error)
	SetForwardingPipelineConfig() error
	Write(req *p4.WriteRequest) <-chan []*p4.Error
	SetWriteTraceChan(traceChan chan WriteTrace)
	ClientId() uint32
	DeviceID() uint32
}

type bfrtClientKey struct {
	host     string
	deviceId uint32
}

type bfrtClient struct {
	client         p4.BfRuntimeClient
	stream         p4.BfRuntime_StreamChannelClient
	clientId       uint32
	deviceId       uint32
	p4Name         string
	writes         chan p4Write
	writeTraceChan chan WriteTrace
	batchSize      int
	numThreads     int
}

func (c *bfrtClient) Init(p4Name string) (err error) {
	c.p4Name = p4Name
	// Initialize stream for mastership and packet I/O
	c.stream, err = c.client.StreamChannel(context.Background())
	if err != nil {
		return
	}
	go func() {
		for {
			_, err := c.stream.Recv()
			if err != nil {
				fmt.Printf("stream recv error: %v\n", err)
			} else {
				fmt.Println("client is master")
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

func (c *bfrtClient) ClientId() uint32 {
	return c.clientId
}

func (c *bfrtClient) DeviceID() uint32 {
	return c.deviceId
}

func CreateOrGetBFRuntimeClient(host string, deviceId uint32, batchSize int, numThreads int, p4Name string) (BFRuntimeClient, error) {

	key := bfrtClientKey{
		host:     host,
		deviceId: deviceId,
	}

	// First, return a P4RT client if one exists
	if p4rtClient, ok := bfrtClients[key]; ok {
		return p4rtClient, nil
	}

	// Second, check to see if we can reuse the gRPC connection for a new P4RT client
	conn, err := GetConnection(host)
	if err != nil {
		return nil, err
	}
	client := &bfrtClient{
		client:     p4.NewBfRuntimeClient(conn),
		deviceId:   deviceId,
		batchSize:  batchSize,
		numThreads: numThreads,
	}
	err = client.Init(p4Name)
	if err != nil {
		return nil, err
	}
	bfrtClients[key] = client
	return client, nil
}
