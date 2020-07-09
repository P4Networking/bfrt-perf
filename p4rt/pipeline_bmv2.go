// +build bmv2
// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0

// To include this file, use go build -tags tofino

package p4rt

import (
	"errors"
	"fmt"
	"os"
)

func LoadDeviceConfig(deviceConfigPath string) (P4DeviceConfig, error) {
	fmt.Printf("BMv2 JSON: %s\n", deviceConfigPath)

	deviceConfig, err := os.Open(deviceConfigPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %v", deviceConfigPath, err)
	}
	defer deviceConfig.Close()
	bmv2BinInfo, err := deviceConfig.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %v", deviceConfigPath, err)
	}

	bin := make([]byte, int(bmv2BinInfo.Size()))
	if b, err := deviceConfig.Read(bin); err != nil {
		return nil, fmt.Errorf("read %s: %v", deviceConfigPath, err)
	} else if b != int(bmv2BinInfo.Size()) {
		return nil, errors.New("bmv2 bin copy failed")
	}

	return bin, nil
}

func TestTarget() string {
	return "bmv2"
}
