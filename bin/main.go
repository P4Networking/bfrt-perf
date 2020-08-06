// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/P4Networking/pisc/util"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/P4Networking/bfrt-perf/bfrt"
	f "github.com/P4Networking/pisc/util/enums"
	"github.com/P4Networking/proto/go/p4"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

var writeReples sync.WaitGroup
var failedWrites uint32
var p4infoHelper bfrt.P4InfoHelper

var (
	target     string
	iterations int
	batchSize  int
	numThreads int
	p4Name     string

	clientId uint32 = 0
	deviceId uint32 = 0
)

func init() {
	flag.StringVar(&target, "target", ":50052", "BFRuntime `<Server IP>:<Server Port>`. By default, :50052")
	flag.IntVar(&iterations, "iterations", 1, "Total iterations to run")
	flag.IntVar(&batchSize, "batchSize", 100, "Number of table entries per batch")
	flag.IntVar(&numThreads, "numThreads", 1, "Number of threads to send write request")
	flag.StringVar(&p4Name, "p4Name", "", "Name of p4 program")
	flag.Parse()
}

func main() {
	client, err := bfrt.CreateOrGetBFRuntimeClient(target, deviceId, batchSize, numThreads, p4Name)
	if err != nil {
		panic(err)
	}

	err = client.SetMastership(clientId)
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	err = client.SetForwardingPipelineConfig()
	if err != nil {
		panic(err)
	}

	bfrtConfig, err := client.GetForwardingPipelineConfig()
	if err != nil {
		panic(err)
	}

	err = p4infoHelper.Init(bfrtConfig[0].BfruntimeInfo)
	if err != nil {
		panic(err)
	}

	// Set up write tracing for test
	writeTraceChan := make(chan bfrt.WriteTrace, 1000)
	client.SetWriteTraceChan(writeTraceChan)
	doneChan := make(chan []time.Duration)
	go func() {
		var currentIteration, lastCount int
		printInterval := 1 * time.Second
		ticker := time.Tick(printInterval)
		durations := make([]time.Duration, iterations)
		for {
			select {
			case trace := <-writeTraceChan:
				durations[currentIteration] = trace.Duration
				currentIteration++
				if currentIteration == iterations {
					doneChan <- durations
					return
				} else if currentIteration > iterations {
					// Should not happened
					panic(fmt.Errorf("Current iteration %d is greater than target iteration number %d", currentIteration, iterations))
				}
			case <-ticker:
				fmt.Printf("\033[2K\rWrote %d of %d (~%.1f flows/sec)...",
					currentIteration, iterations, float64(currentIteration-lastCount)/printInterval.Seconds())
				lastCount = currentIteration
			}
		}
	}()

	// Send the flow entries
	writeReples.Add(iterations)
	SendTableEntries(client, iterations, batchSize)

	// Wait for all writes to finish
	durations := <-doneChan
	writeReples.Wait()
	fmt.Printf("Number of failed writes: %d\n", failedWrites)

	// Writing to CSV file
	fileName := fmt.Sprintf("test-result-%s-%d-%d-%d.csv", "Tofino", batchSize, iterations, time.Now().Unix())
	fmt.Printf("Saving results to %s\n", fileName)

	csvFile, err := os.Create(fileName)
	defer csvFile.Close()
	if err != nil {
		panic(err)
	}
	resultWriter := csv.NewWriter(csvFile)

	resultWriter.Write([]string{"Index of durations", "µs/per write request"})
	var summary int64
	for i, d := range durations {
		data := []string{strconv.Itoa(i), strconv.FormatInt(d.Microseconds(), 10)}
		resultWriter.Write(data)
		summary += d.Microseconds()
	}
	resultWriter.Flush()
	fmt.Printf("\033[2K\r%f seconds, %d writes, %f writes request/sec\n",
		float64(summary)/1000000, iterations, float64(int64(iterations)*1000000)/float64(summary))
}

// SendTableEntries writes multiple table entries to the routing_v4
// table.
func SendTableEntries(client bfrt.BFRuntimeClient, iterations int, batchSize int) {

	tableID, err := p4infoHelper.GetP4Id("pipe.SwitchIngress.rib_24")
	if err != nil {
		panic(err)
	}

	actionID, err := p4infoHelper.GetP4Id("SwitchIngress.hit_route_port")
	if err != nil {
		panic(err)
	}

	// Prepare write requests for all iterations
	requests := make([]*p4.WriteRequest, iterations)
	for i := 0; i < iterations; i++ {
		updates := make([]*p4.Update, batchSize)
		for j := 0; j < batchSize; j++ {
			ipOri := int2ip(uint32((i*batchSize + j)))
			ip := net.IPv4(ipOri[1], ipOri[2], ipOri[3], ipOri[0])

			updates[j] = &p4.Update{
				Type: p4.Update_INSERT,
				Entity: &p4.Entity{Entity: &p4.Entity_TableEntry{
					TableEntry: &p4.TableEntry{
						TableId: tableID, // FabricIngress.forwarding.routing_v4
						Key: &p4.TableKey{
							Fields: []*p4.KeyField{
								util.GenKeyField(f.MATCH_EXACT,
									1,
									util.Ipv4ToBytes(ip.String())[0:3]),
							},
						},
						Data: &p4.TableData{
							ActionId: actionID,
							Fields: []*p4.DataField{
								util.GenDataField(1, util.Int16ToBytes(128)),
							},
						},
					},
				},
				},
			}
		}
		req := &p4.WriteRequest{
			ClientId: client.ClientId(),
			Target: &p4.TargetDevice{
				DeviceId: client.DeviceID(),
				PipeId:   0xffff,
			},
			Updates: updates,
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

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}
