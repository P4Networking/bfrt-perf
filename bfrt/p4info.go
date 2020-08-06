// Copyright 2020-present Brian O'Connor
// Copyright 2020-present Open Networking Foundation
// SPDX-License-Identifier: Apache-2.0
// Modifications copyright (C) 2020 Chun-Ming Ou

package bfrt

import (
	"encoding/json"
	"fmt"
	"github.com/P4Networking/pisc/util"
)

type P4InfoHelper struct {
	nameToP4ID map[string]uint32 // P4 name to P4 ID.

}

func (p4infoHelper *P4InfoHelper) Init(config []byte) (err error) {
	err = json.Unmarshal(config, &util.BfRtInfo_P4)
	if err != nil {
		return err
	}

	p4infoHelper.nameToP4ID = make(map[string]uint32)

	for _, table := range util.BfRtInfo_P4.Tables {
		p4infoHelper.nameToP4ID[table.Name] = table.ID
		for _, action := range table.ActionSpecs {
			p4infoHelper.nameToP4ID[action.Name] = uint32(action.ID)
		}
	}
	return
}

func (p4infoHelper *P4InfoHelper) GetP4Id(name string) (p4ID uint32, err error) {
	p4ID, exists := p4infoHelper.nameToP4ID[name]
	if !exists {
		err = fmt.Errorf("Unable to find P4 ID for %s", name)
	}
	return
}
