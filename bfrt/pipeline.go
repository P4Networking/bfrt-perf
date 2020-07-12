// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package bfrt

import (
	"context"
	p4 "github.com/breezestars/go-bfrt/proto/out"
	"github.com/pkg/errors"
)

func getPipelineConfig(client p4.BfRuntimeClient, clientId, deviceId uint32) ([]*p4.ForwardingPipelineConfig, error) {
	req := &p4.GetForwardingPipelineConfigRequest{
		ClientId: clientId,
		DeviceId: deviceId,
	}
	res, err := client.GetForwardingPipelineConfig(context.Background(), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting pipeline config")
	}
	res.GetConfig()
	return res.GetConfig(), nil
}

func setPipelineConfig(client p4.BfRuntimeClient, clientId, deviceId uint32, p4Name string) error {
	req := &p4.SetForwardingPipelineConfigRequest{
		ClientId: clientId,
		DeviceId: deviceId,
		Action:   p4.SetForwardingPipelineConfigRequest_BIND,
		Config: []*p4.ForwardingPipelineConfig{
			{
				P4Name: p4Name,
			},
		},
	}
	_, err := client.SetForwardingPipelineConfig(context.Background(), req)
	// ignore the response; it is an empty message
	return err
}

func (c *bfrtClient) SetForwardingPipelineConfig() (err error) {
	err = setPipelineConfig(c.client, c.clientId, c.deviceId, c.p4Name)
	if err != nil {
		return
	}
	return
}

func (c *bfrtClient) GetForwardingPipelineConfig() ([]*p4.ForwardingPipelineConfig, error) {
	return getPipelineConfig(c.client, c.clientId, c.deviceId)
}
