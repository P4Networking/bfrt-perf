// +build stratum-bfrt
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
	deviceConfigBin, err := os.Open(deviceConfigPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %v", deviceConfigPath, err)
	}
	defer deviceConfigBin.Close()
	deviceConfigBinInfo, err := deviceConfigBin.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %v", deviceConfigPath, err)
	}

	// Allocate the device config buffer
	binLen := int(deviceConfigBinInfo.Size())
	bin := make([]byte, binLen)

	if bytesRead, err := deviceConfigBin.Read(bin); err != nil {
		return nil, fmt.Errorf("read %s: %v", tofinoBinPath, err)
	} else if bytesRead != int(tofinoBinInfo.Size()) {
		return nil, errors.New("tofino bin copy failed")
	}

	return bin, nil
}
