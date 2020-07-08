// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/binary"
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
	verbose := flag.Bool("verbose", false, "")
	p4infoPath := flag.String("p4info", "", "")
	count := flag.Uint64("count", 1, "")
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
	writeTraceChan := make(chan p4rt.WriteTrace, 100)
	client.SetWriteTraceChan(writeTraceChan)
	doneChan := make(chan bool)
	go func() {
		var writeCount, lastCount uint64
		printInterval := 1 * time.Second
		ticker := time.Tick(printInterval)
		for {
			select {
			case trace := <-writeTraceChan:
				writeCount += uint64(trace.BatchSize)
				if writeCount >= *count {
					doneChan <- true
					return
				}
			case <-ticker:
				if *verbose {
					fmt.Printf("\033[2K\rWrote %d of %d (~%.1f flows/sec)...",
						writeCount, *count, float64(writeCount-lastCount)/printInterval.Seconds())
					lastCount = writeCount
				}
			}
		}
	}()

	// Send the flow entries
	writeReples.Add(int(*count))
	start := time.Now()
	SendTableEntries(client, *count)

	// Wait for all writes to finish
	<-doneChan
	duration := time.Since(start).Seconds()
	fmt.Printf("\033[2K\r%f seconds, %d writes, %f writes/sec\n",
		duration, *count, float64(*count)/duration)
	writeReples.Wait()
	fmt.Printf("Number of failed writes: %d\n", failedWrites)
}

// SendTableEntries writes multiple table entries to the routing_v4
// table.
func SendTableEntries(p4rt p4rt.P4RuntimeClient, count uint64) {
	match := []*p4.FieldMatch{
		{
			FieldId:        1, // ipv4_dst
			FieldMatchType: &p4.FieldMatch_Lpm{&p4.FieldMatch_LPM{}},
		},
	}

	routeV4TableID, err := p4infoHelper.GetP4Id("FabricIngress.forwarding.routing_v4")
	if err != nil {
		panic(err)
	}

	setNextActionID, err := p4infoHelper.GetP4Id("FabricIngress.forwarding.set_next_id_routing_v4")
	if err != nil {
		panic(err)
	}

	update := &p4.Update{
		Type: p4.Update_INSERT,
		Entity: &p4.Entity{Entity: &p4.Entity_TableEntry{
			TableEntry: &p4.TableEntry{
				TableId: routeV4TableID, // FabricIngress.forwarding.routing_v4
				Match:   match,
				Action: &p4.TableAction{Type: &p4.TableAction_Action{Action: &p4.Action{
					ActionId: setNextActionID, // set_next_id_routing_v4
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

	for i := uint64(0); i < count; i++ {
		matchField := update.GetEntity().GetTableEntry().GetMatch()[0].GetLpm()
		matchField.Value = Uint64(i)[4:8] // ipv4 is 32 bits
		matchField.PrefixLen = 32
		res := p4rt.Write(update)
		go CountFailed(proto.Clone(update).(*p4.Update), res)
	}
}

func CountFailed(update *p4.Update, res <-chan *p4.Error) {
	err := <-res
	if err.CanonicalCode != int32(codes.OK) { // write failed
		atomic.AddUint32(&failedWrites, 1)
		fmt.Fprintf(os.Stderr, "%v -> %v\n", update, err.GetMessage())
	}
	writeReples.Done()
}

func Uint64(v uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	return bytes
}
