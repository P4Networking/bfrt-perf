// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	p4 "github.com/p4lang/p4runtime/go/p4/v1"
)

func (c *p4rtClient) SetMastership(electionId p4.Uint128) (err error) {
	c.electionId = electionId
	mastershipReq := &p4.StreamMessageRequest{
		Update: &p4.StreamMessageRequest_Arbitration{
			Arbitration: &p4.MasterArbitrationUpdate{
				DeviceId:   1,
				ElectionId: &electionId,
			},
		},
	}
	err = c.stream.Send(mastershipReq)
	return
}
