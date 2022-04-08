package main

import (
	"fmt"
	ygotsrl "steiler/yangtest/generated/srl"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

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
