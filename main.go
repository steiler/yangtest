package main

import (
	"fmt"

	ygotsrl "steiler/yangtest/generated/srl"
	"steiler/yangtest/generated/sros"

	"github.com/openconfig/ygot/ygot"
)

func main() {
	srosDevice := &ygotsros.Device{}
}

func old_main() {
	// First lets prepare some stuff.
	// retieve a simple very small config, which we consider the actual config
	// this would be cached in the controller and probably come from an array, in which all the configs are stored / cached.
	actualConfig := &ygotsrl.Device{}
	//actualConfig := loadConfigFromFile("/home/steiler/projects/yangtest/configwim2 init.json")

	// // Uncomment to add ethernet-1-50 description and stuff to the actual config, this is to check for the MergeOverwrite functionality
	// appendE150(actualConfig)

	// retrieve a config snippet defining a subinterface as well as the network-instance as default for the /system/ssh-server
	// this would be the spec with which the controller would be triggered
	//specConfig := loadConfigFromFile("/home/steiler/projects/yangtest/configwim2 copy.json")
	specConfig := loadConfigFromFile("/home/steiler/projects/yangtest/configwim3.json")

	// lets start our "fake" reconsiliation
	DoComparisonAndPathEvaluation(actualConfig, specConfig)
}

// this is where we drive the whole stuff,
// basically parts of the reconsiliation
func DoComparisonAndPathEvaluation(actualConfig *ygotsrl.Device, specConfig *ygotsrl.Device) {

	// validate the actual config
	err := actualConfig.Validate()
	if err != nil {
		panic(err)
	}

	// skipping specValidation, this will probably result in missing leaf leafrefs
	newConfigTmp, err := ygot.DeepCopy(actualConfig)
	if err != nil {
		panic(err)
	}
	newConfig := newConfigTmp.(*ygotsrl.Device) // Typecast

	// Merge spec into newconfig, which is right now jsut the actual config
	err = ygot.MergeStructInto(newConfig, specConfig)
	if err != nil {
		panic(err)
	}

	// validate the new config
	err = newConfig.Validate()
	if err != nil {
		panic(err)
	}

	// create a diff of the actual compared to the to-become-new config
	actualVsSpecDiff, err := ygot.Diff(actualConfig, newConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("GNMI Notification Content:")
	printGnmiNotification(actualVsSpecDiff)
	fmt.Println("")

	// calculate the relevant paths (As wim calls them, Root Paths)
	//_ = CarveOutRelevantSubPaths(actualVsSpecDiff)

	fmt.Println("Relevant Hier-Paths:")

	// get Root Schema
	// schema, err := ygotsrl.Schema()
	// if err != nil {
	// 	panic(err)
	// }
	// deviceSchema := schema.RootSchema()

	// Feed schema and diff *gnmi.Update into rootpath library
	//rootConfigElement := rootpaths.ConfigElementHierarchyFromGnmiUpdate(deviceSchema, actualVsSpecDiff)

	// // get and print Rootpaths
	// printGnmiPaths(rootConfigElement.GetRootPaths())

	// fmt.Println(rootConfigElement.GetHierarchicalOutput(" "))

	fmt.Println("All good, made it to the end!")
}

/* // Carve out the sub-paths that have a specific relevance from a gnmi.Notification.
//
// We go down the tree to find the paths where values are actually changed or deleted
// and returns a list of these gnmi.Path's in an aggregated way.
// Or put it, returns the absolute path to the root of the elements which will be touched in the gnmi.Notification.
func CarveOutRelevantSubPaths(gn *gnmi.Notification) []*gnmi.Path {
	//retrieve the root schema which is required in the following
	rootSchema := getRootSchema()

	// initialize storage for all the relevant paths
	allPathSchemas := []*PathAndSchema{}

	for _, elem := range gn.GetUpdate() {
		// for each update check that it is no default value, we want to skip those.
		pathAndSchemaElem := getPathAndSchemaEntry(rootSchema, elem.Path)
		allPathSchemas = append(allPathSchemas, pathAndSchemaElem)

		runhierarchical(pathAndSchemaElem)
	}

	fmt.Println(rootCE.getHierarchicalOutput(0, "  "))

	// aggregate the deduced paths to find the common relevant base paths
	AggregateSpecSignificantPaths := aggregateCommonPaths(allPathSchemas, rootSchema)

	for _, ps := range AggregateSpecSignificantPaths {
		fmt.Println(ps.String())
	}

	return append(gn.GetDelete())
}
*/

/* // checks if the provided gnmi.Update just sets the key for the corresponding entry or any other related data.
func isKeyValue(schemaEntry *yang.Entry, u *gnmi.Update) bool {
	return schemaEntry.Parent.Key == u.Path.Elem[len(u.Path.Elem)-1].Name
}

// checks if the provided gnmi.Update just sets a value to the default value of a leaf
func isDefaultValue(schemaEntry *yang.Entry, u *gnmi.Update) bool {
	if defval, singleDefVal := schemaEntry.SingleDefaultValue(); singleDefVal {
		return u.Val.GetStringVal() == defval
	}
	return false
} */
/*
// aggregates the proivided paths by extratcting relevant paths
func aggregateCommonPaths(p []*PathAndSchema, rootSchema *yang.Entry) []*PathAndSchemaCount {
	// we add the first path straight away to the list, therrefore make sure we have at least 1 entry in the list
	result := []*PathAndSchemaCount{}
	if len(p) == 0 {
		return result
	}
	// append the first struct to the result slice
	result = append(result, &PathAndSchemaCount{p[0], 1})
	var found bool

	// iterate over all paths
	for _, elem := range p[1 : len(p)-1] {
		found = false
		// compare to existing paths in result
		for resultIndex, resultElem := range result {
			// extract the common path of the actual elem and the path from the actual result paths entry
			cp := ygotutils.FindPathElemPrefix([]*gnmi.Path{elem.path, resultElem.path})
			if cp != nil && len(cp.Elem) > 0 {
				newPath := &gnmi.Path{Elem: cp.Elem}
				// create a new PathAndSchema struct taking the previouse count incremented by 1
				result[resultIndex] = &PathAndSchemaCount{getPathAndSchemaEntry(rootSchema, newPath), resultElem.count + 1}
				found = true
				break
			}
		}
		if !found {
			// there was no match for such path so we add it to the result with a count of 1
			result = append(result, &PathAndSchemaCount{elem, 1})
		}
	}
	return result
} */
