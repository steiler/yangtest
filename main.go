package main

import (
	"fmt"
	ygotsrl "steiler/yangtest/generated"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

func main() {

	// First lets prepare some stuff.
	// retieve a simple very small config, which we consider the actual config
	// this would be cached in the controller and probably come from an array, in which all the configs are stored / cached.
	actualConfig := getActual()
	// retrieve a config snippet defining a subinterface as well as the network-instance as default for the /system/ssh-server
	// this would be the spec with which the controller would be triggered
	specConfig := getSpec()

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
	relevantPaths := CarveOutRelevantSubPaths(actualVsSpecDiff)

	fmt.Println("Relevant Paths:")
	printGnmiPaths(relevantPaths)
	fmt.Println("")

	fmt.Println("All good, made it to the end!")

}

// Carve out the sub-paths that have a specific relevance from a gnmi.Notification.
//
// We go down the tree to find the paths where values are actually changed or deleted
// and returns a list of these gnmi.Path's in an aggregated way.
// Or put it, returns the absolute path to the root of the elements which will be touched in the gnmi.Notification.
func CarveOutRelevantSubPaths(gn *gnmi.Notification) []*gnmi.Path {

	allSpecSignificantPaths := []*gnmi.Path{}

	// Deletes are all significant so we add them straight
	allSpecSignificantPaths = append(allSpecSignificantPaths, gn.GetDelete()...)

	for _, elem := range gn.GetUpdate() {
		allSpecSignificantPaths = append(allSpecSignificantPaths, elem.Path)
	}

	// aggregate the deduced paths to find the common relevant base paths
	AggregateSpecSignificantPaths := aggregateCommonPaths(allSpecSignificantPaths)

	return AggregateSpecSignificantPaths
}

/* func referencesOnlyKeyInList(gp *gnmi.Path) bool {
	if len(gp.Elem) > 1 {
		length := len(gp.Elem)
		lastElem := gp.Elem[length-1]
		priorToLastElem := gp.Elem[length-2]
		_, exists := priorToLastElem.Key[lastElem.Name]
		return exists
	}
	return false
} */

//
func aggregateCommonPaths(p []*gnmi.Path) []*gnmi.Path {
	result := []*gnmi.Path{}
	if len(p) == 0 {
		return result
	}
	result = append(result, p[0])
	var found bool

	// iterate over all paths
	for _, elem := range p {
		found = false
		// compare to existing paths in result
		for resultIndex, resultElem := range result {
			// extract the common path of the actual elem and the path from the actual result paths entry
			cp := extractCommonPath(elem, resultElem)
			if len(cp) > 0 {
				result[resultIndex] = &gnmi.Path{Elem: cp}
				found = true
				break
			}
		}
		if !found {
			result = append(result, elem)
		}
	}
	return result
}

// extract the share part of the two provided gnmi.Paths
// if the two provided paths are fully disjoint an empty array is returned
// otherwise the comon part of the path is returned
func extractCommonPath(a, b *gnmi.Path) []*gnmi.PathElem {
	minElems := len(a.Elem)
	if len(b.Elem) < minElems {
		minElems = len(b.Elem)
	}
	result := []*gnmi.PathElem{}

	for i := 0; i < minElems; i++ {
		if pathElemIsEqual(a.Elem[i], b.Elem[i]) {
			result = append(result, a.Elem[i])
		} else {
			break
		}
	}
	return result
}

// check for the Name, Key and Value based equality of two gnmi.PathElement structs
func pathElemIsEqual(a, b *gnmi.PathElem) bool {
	if a.Name != b.Name {
		// Element names are already non-equal
		return false
	}
	var bValue string
	var bKeyExists bool
	// check for the key names and key values equallity
	// iterate over a keys
	for aKey, aValue := range a.Key {
		// if key exists check key value
		if bValue, bKeyExists = b.Key[aKey]; !bKeyExists {
			return false
		}
		// make sure the values are equal
		if aValue != bValue {
			return false
		}
	}
	return true
}
