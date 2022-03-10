package main

import (
	"fmt"
	ygotsrl "steiler/yangtest/generated"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

func getActual() *ygotsrl.Device {
	config := `{
		"interface": [
		   {
			  "admin-state": "enable",
			  "description": "if1",
			  "name": "ethernet-1/1",
			  "subinterface": [
				 {
					"index": 1,
					"ipv4": {
					   "address": [
						  {
							 "ip-prefix": "192.168.1.1/24"
						  }
					   ]
					}
				 }
			  ]
		   },
		   {
			  "admin-state": "enable",
			  "description": "if2",
			  "name": "ethernet-1/2",
			  "subinterface": [
				 {
					"index": 1,
					"ipv4": {
					   "address": [
						  {
							 "ip-prefix": "192.168.2.1/24"
						  }
					   ]
					}
				 }
			  ]
		   }
		]
	 }`

	dx := &ygotsrl.Device{}

	ygotsrl.Unmarshal([]byte(config), dx)
	dx.GetOrCreateNetworkInstance("default")

	return dx
}

func getSpec() *ygotsrl.Device {

	specDevice := &ygotsrl.Device{}

	if1 := specDevice.GetOrCreateInterface("ethernet-1/1")

	si, err := createSubInterfaceWithIP(5, "192.168.200.200/24")
	if err != nil {
		panic(err)
	}

	if1.AppendSubinterface(si)

	sif3, err := if1.NewSubinterface(46)
	if err != nil {
		panic(err)
	}
	vlan := sif3.GetOrCreateVlan()

	enc := vlan.GetOrCreateEncap() //.SingleTagged.VlanId = ygotsrl.UnionUint16(5)
	enc.GetOrCreateSingleTagged().VlanId = ygotsrl.UnionUint16(5)

	sif3.GetOrCreateIpv4().NewAddress("192.168.5.1/24")

	specDevice.GetOrCreateSystem().GetOrCreateSshServer().GetOrCreateNetworkInstance("default").AdminState = ygotsrl.SrlNokiaCommon_AdminState_enable

	return specDevice
}

func printGnmiPaths(p []*gnmi.Path) {
	for _, rp := range p {
		fmt.Println(rp.String())
	}

}

func createInterfaceWithIP(name string, description string, subIfIndex uint32, ip string) (*ygotsrl.SrlNokiaInterfaces_Interface, error) {
	interf := &ygotsrl.SrlNokiaInterfaces_Interface{Name: ygot.String(name), Description: ygot.String(description)}
	suif, err := createSubInterfaceWithIP(subIfIndex, ip)
	if err != nil {
		return nil, err
	}
	interf.AppendSubinterface(suif)
	return interf, nil

}
