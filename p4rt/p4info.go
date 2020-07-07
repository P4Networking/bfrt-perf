package p4rt

import (
	"fmt"
	"io/ioutil"
	"github.com/golang/protobuf/proto"

	p4_config "github.com/p4lang/p4runtime/go/p4/config/v1"
)

type P4InfoHelper struct {
	nameToP4ID	map[string]uint32  // P4 name to P4 ID.

}


func LoadP4Info(p4infoPath string) (p4info p4_config.P4Info, err error) {
	fmt.Printf("P4 Info: %s\n", p4infoPath)

	p4infoBytes, err := ioutil.ReadFile(p4infoPath)
	if err != nil {
		return
	}
	err = proto.UnmarshalText(string(p4infoBytes), &p4info)
	return
}

func (p4infoHelper *P4InfoHelper) Init(p4InfoPath string) (err error) {
	var p4info	p4_config.P4Info
	p4infoHelper.nameToP4ID = make(map[string]uint32)
	p4info, err = LoadP4Info(p4InfoPath)
	if err != nil {
		return
	}

	for _, table := range p4info.Tables {
		p4infoHelper.nameToP4ID[table.GetPreamble().Name] = table.GetPreamble().Id
	}

	for _, action := range p4info.Actions {
		p4infoHelper.nameToP4ID[action.GetPreamble().GetName()] = action.GetPreamble().GetId()
	}
	return
}

func (p4infoHelper *P4InfoHelper) GetP4Id(name string) (p4ID uint32, err error){
	p4ID, exists := p4infoHelper.nameToP4ID[name]
	if !exists {
		err = fmt.Errorf("Unable to find P4 ID for %s", name)
	}
	return
}
