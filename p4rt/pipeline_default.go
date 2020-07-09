// +build !bmv2,!stratum_bf,!stratum_bfrt
// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

// This file is included in no build tag specifies the target platform

package p4rt

import (
	"errors"
)

func LoadDeviceConfig(deviceConfigPath string) (P4DeviceConfig, error) {
	return nil, errors.New("No target type specified at build time. " +
		"You need to rebuild with \"-tags\"")
}

func TestTarget() string {
	return "default"
}
