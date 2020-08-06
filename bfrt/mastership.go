// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package bfrt

import (
	"github.com/P4Networking/proto/go/p4"
)

func (c *bfrtClient) SetMastership(clientId uint32) (err error) {
	c.clientId = clientId

	mastershipReq := &p4.StreamMessageRequest{
		Update: &p4.StreamMessageRequest_Subscribe{
			Subscribe: &p4.Subscribe{
				DeviceId: c.deviceId,
				IsMaster: true,
				Notifications: &p4.Subscribe_Notifications{
					EnableLearnNotifications:            true,
					EnableIdletimeoutNotifications:      true,
					EnablePortStatusChangeNotifications: true,
				},
			},
		},
	}

	err = c.stream.Send(mastershipReq)
	return
}
