package main

import (
	"fmt"
	ygotsrl "steiler/yangtest/generated"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

func OtherExample() {

	d1 := &ygotsrl.Device{}

	d2 := &ygotsrl.Device{}

	// if31, err := d1.NewInterface("ethernet-1/3")
	// if err != nil {
	// 	panic(err)
	// }
	// _ = if31

	if3, err := d2.NewInterface("ethernet-1/3")
	if err != nil {
		panic(err)
	}
	if3.AdminState = ygotsrl.SrlNokiaCommon_AdminState_enable
	if3.VlanTagging = ygot.Bool(true)

	sif3, err := if3.NewSubinterface(5)
	if err != nil {
		panic(err)
	}
	vlan := sif3.GetOrCreateVlan()

	enc := vlan.GetOrCreateEncap() //.SingleTagged.VlanId = ygotsrl.UnionUint16(5)
	enc.GetOrCreateSingleTagged().VlanId = ygotsrl.UnionUint16(5)

	sif3.GetOrCreateIpv4().NewAddress("192.168.3.1/24")

	gnmin, err := ygot.Diff(d1, d2)
	if err != nil {
		panic(err)
	}

	_ = gnmin

	printGnmiNotification(gnmin)
	fmt.Println("")
	fmt.Println("xxxxxxxxxxxxx")
	fmt.Println("")
	relevantPaths := CarveOutRelevantSubPaths(gnmin)

	for _, rp := range relevantPaths {
		fmt.Println(rp.String())
	}

	err = d1.Validate()
	if err != nil {
		panic(err)
	}
	err = d2.Validate()
	if err != nil {
		panic(err)
	}

	err = ygot.MergeStructInto(d2, d1)
	if err != nil {
		panic(err)
	}

	// ################################################

	printSRL(d1)

}

func Test1() {
	d := &ygotsrl.Device{}

	json := ""

	if err := ygotsrl.Unmarshal([]byte(json), d); err != nil {
		panic(fmt.Sprintf("Cannot unmarshal JSON: %v", err))
	}
}

func buildConfig2() *ygotsrl.Device {
	// create a device
	d := &ygotsrl.Device{}

	// retrieve a new interface with a single subinterface and an ip
	if1, err := createInterfaceWithIP("ethernet-1/3", "if1", 5, "192.168.3.1/24")
	if err != nil {
		panic(err)
	}

	// append the interfce to the device
	d.AppendInterface(if1)

	return d
}

func buildConfig1() *ygotsrl.Device {
	// create a device
	d := &ygotsrl.Device{}

	// retrieve a new interface with a single subinterface and an ip
	if1, err := createInterfaceWithIP("ethernet-1/1", "if1", 5, "192.168.1.1/24")
	if err != nil {
		panic(err)
	}

	// append the interfce to the device
	d.AppendInterface(if1)

	// retrieve a new interface with a single subinterface and an ip
	if2, err := createInterfaceWithIP("ethernet-1/2", "if2", 5, "192.168.2.1/24")
	if err != nil {
		panic(err)
	}

	// append the interfce to the device
	d.AppendInterface(if2)

	return d
}

func printSRL(d *ygotsrl.Device) {
	s, err := ygot.EmitJSON(d, &ygot.EmitJSONConfig{Format: ygot.RFC7951})
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}

func createSubInterfaceWithIP(index uint32, ip string) (*ygotsrl.SrlNokiaInterfaces_Interface_Subinterface, error) {

	si1 := &ygotsrl.SrlNokiaInterfaces_Interface_Subinterface{
		Index: ygot.Uint32(index),
	}

	ipv4 := si1.GetOrCreateIpv4()
	_, err := ipv4.NewAddress(ip)

	if err != nil {
		panic(err)
	}
	return si1, nil
}

func printGnmiNotification(gn *gnmi.Notification) {
	for x, elem := range gn.GetDelete() {
		fmt.Printf("Delete %d: %s\n", x, elem.String())
	}
	for x, elem := range gn.GetUpdate() {
		fmt.Printf("Update %d: %s\n", x, elem.String())
	}
}

func configFromString() *ygotsrl.Device {
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

	return dx
}

func complexMergeDiff() {
	d1 := configFromString()
	d2tmp, err := ygot.DeepCopy(d1)
	d2 := d2tmp.(*ygotsrl.Device)

	if err != nil {
		panic(err)
	}
	if3, err := d2.NewInterface("ethernet-1/3")
	if err != nil {
		panic(err)
	}

	if3.VlanTagging = ygot.Bool(true)

	sif3, err := if3.NewSubinterface(5)
	if err != nil {
		panic(err)
	}
	vlan := sif3.GetOrCreateVlan()

	enc := vlan.GetOrCreateEncap() //.SingleTagged.VlanId = ygotsrl.UnionUint16(5)
	enc.GetOrCreateSingleTagged().VlanId = ygotsrl.UnionUint16(5)

	sif3.GetOrCreateIpv4().NewAddress("192.168.3.1/24")

	gnmin, err := ygot.Diff(d2, d1)
	if err != nil {
		panic(err)
	}

	_ = gnmin

	printGnmiNotification(gnmin)

	relevantPaths := CarveOutRelevantSubPaths(gnmin)

	for _, rp := range relevantPaths {
		fmt.Println(rp.String())
	}

	err = d1.Validate()
	if err != nil {
		panic(err)
	}
	err = d2.Validate()
	if err != nil {
		panic(err)
	}

	err = ygot.MergeStructInto(d1, d2)
	if err != nil {
		panic(err)
	}

	// ################################################

	printSRL(d1)
}
