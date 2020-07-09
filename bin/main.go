// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Yi-Tseng/p4r-perf/p4rt"
	"github.com/golang/protobuf/proto"
	p4 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc/codes"
)

var writeReples sync.WaitGroup
var failedWrites uint32
var p4infoHelper p4rt.P4InfoHelper

func main() {

	target := flag.String("target", "localhost:28000", "")
	p4infoPath := flag.String("p4info", "", "")
	iterations := flag.Int("iterations", 1, "total iterations to run")
	deviceConfig := flag.String("deviceConfig", "", "")
	batchSize := flag.Int("batchSize", 100, "Number of table entries per batch.")
	numThreads := flag.Int("numThreads", 1, "Number of threads to send write request.")

	flag.Parse()

	client, err := p4rt.CreateOrGetP4RuntimeClient(*target, 1, *batchSize, *numThreads)
	if err != nil {
		panic(err)
	}

	err = client.SetMastership(p4.Uint128{High: 0, Low: 1})
	if err != nil {
		panic(err)
	}

	err = client.SetForwardingPipelineConfig(*p4infoPath, *deviceConfig)
	if err != nil {
		panic(err)
	}

	err = p4infoHelper.Init(*p4infoPath)
	if err != nil {
		panic(err)
	}

	// Set up write tracing for test
	writeTraceChan := make(chan p4rt.WriteTrace, 1000)
	client.SetWriteTraceChan(writeTraceChan)
	doneChan := make(chan []time.Duration)
	go func() {
		var currentIteration int
		durations := make([]time.Duration, *iterations)
		for {
			select {
			case trace := <-writeTraceChan:
				durations[currentIteration] = trace.Duration
				currentIteration++
				if currentIteration == *iterations {
					doneChan <- durations
					return
				} else if currentIteration > *iterations {
					// Should not happened
					panic(fmt.Errorf("Current iteration %d is greater than target iteration number %d", currentIteration, *iterations))
				}
			}
		}
	}()

	// Send the flow entries
	writeReples.Add(int(*iterations))
	SendTableEntries(client, *iterations, *batchSize)

	// Wait for all writes to finish
	durations := <-doneChan

	writeReples.Wait()
	fmt.Printf("Number of failed writes: %d\n", failedWrites)

	csvFile, err := os.Create("result.csv")
	defer csvFile.Close()
	if err != nil {
		panic(err)
	}
	resultWriter := csv.NewWriter(csvFile)

	for i, d := range durations {
		data := []string{string(i), string(d.Microseconds())}
		resultWriter.Write(data)
	}
	resultWriter.Flush()
}

// SendTableEntries writes multiple table entries to the routing_v4
// table.
func SendTableEntries(client p4rt.P4RuntimeClient, iterations int, batchSize int) {

	// Prepare write requests for all iterations
	requests := make([]*p4.WriteRequest, iterations)
	for i := 0; i < iterations; i++ {
		tableID, err := p4infoHelper.GetP4Id("FabricIngress.forwarding.routing_v4")
		if err != nil {
			panic(err)
		}

		actionID, err := p4infoHelper.GetP4Id("FabricIngress.forwarding.set_next_id_routing_v4")
		if err != nil {
			panic(err)
		}

		updates := make([]*p4.Update, batchSize)
		for j := 0; j < batchSize; j++ {
			match := []*p4.FieldMatch{
				{
					FieldId: 1, // ipv4_dst
					FieldMatchType: &p4.FieldMatch_Lpm{
						&p4.FieldMatch_LPM{
							Value:     Uint64(uint64(i*batchSize + j))[4:8],
							PrefixLen: 32,
						},
					},
				},
			}

			updates[j] = &p4.Update{
				Type: p4.Update_INSERT,
				Entity: &p4.Entity{Entity: &p4.Entity_TableEntry{
					TableEntry: &p4.TableEntry{
						TableId: tableID, // FabricIngress.forwarding.routing_v4
						Match:   match,
						Action: &p4.TableAction{Type: &p4.TableAction_Action{Action: &p4.Action{
							ActionId: actionID, // set_next_id_routing_v4
							Params: []*p4.Action_Param{
								{
									ParamId: 1,              // next_id
									Value:   Uint64(1)[4:8], // 32 bits
								},
							},
						}}},
					},
				}},
			}
		}
		req := &p4.WriteRequest{
			DeviceId:   client.DeviceID(),
			ElectionId: client.ElectionID(),
			Updates:    updates,
		}
		requests[i] = req
	}

	for _, req := range requests {
		res := client.Write(req)
		go CountFailed(proto.Clone(req).(*p4.WriteRequest), res)
	}
}

func CountFailed(write *p4.WriteRequest, res <-chan []*p4.Error) {
	errors := <-res
	for i, err := range errors {
		update := write.Updates[i]
		if err.CanonicalCode != int32(codes.OK) { // write failed
			atomic.AddUint32(&failedWrites, 1)
			fmt.Fprintf(os.Stderr, "%v -> %v\n", update, err.GetMessage())
		}
	}
	writeReples.Done()
}

func Uint64(v uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	return bytes
}
