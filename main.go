package main

import (
	"fmt"
	ygotsrl "steiler/yangtest/generated"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
)

func main() {

	// First lets prepare some stuff.
	// retieve a simple very small config, which we consider the actual config
	// this would be cached in the controller and probably come from an array, in which all the configs are stored / cached.
	actualConfig := &ygotsrl.Device{}

	// // Uncomment to add ethernet-1-50 description and stuff to the actual config, this is to check for the MergeOverwrite functionality
	// appendE150(actualConfig)

	// retrieve a config snippet defining a subinterface as well as the network-instance as default for the /system/ssh-server
	// this would be the spec with which the controller would be triggered
	specConfig := loadConfigFromFile("/home/steiler/projects/yangtest/configwim.json")

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
	//retrieve the root schema which is required in the following
	rootSchema := getRootSchema()

	// initialize storage for all the relevant paths
	allSpecSignificantPaths := []*gnmi.Path{}

	// Deletes are all significant so we add them straight
	allSpecSignificantPaths = append(allSpecSignificantPaths, gn.GetDelete()...)

	for _, elem := range gn.GetUpdate() {
		// for each update check that it is no default value, we want to skip those.
		schemaEntry := getSchemaEntry(rootSchema, elem)
		// make sure we are not addind paths with default values or path which are keys
		if !isDefaultValue(schemaEntry, elem) && !isKeyValue(schemaEntry, elem) {
			allSpecSignificantPaths = append(allSpecSignificantPaths, elem.Path)
		}
	}

	// aggregate the deduced paths to find the common relevant base paths
	AggregateSpecSignificantPaths := aggregateCommonPaths(allSpecSignificantPaths)

	return AggregateSpecSignificantPaths
}

// runs from schema root through to the schema element that the gnmi.Update referes to
// and returns the corresponding *yang.Entry.
func getSchemaEntry(rootschema *yang.Entry, u *gnmi.Update) *yang.Entry {
	var schema = rootschema
	for _, elem := range u.Path.Elem {
		schema = schema.Dir[elem.Name]
	}
	return schema
}

// checks if the provided gnmi.Update just sets the key for the corresponding entry or any other related data.
func isKeyValue(schemaEntry *yang.Entry, u *gnmi.Update) bool {
	return schemaEntry.Parent.Key == u.Path.Elem[len(u.Path.Elem)-1].Name
}

// checks if the provided gnmi.Update just sets a value to the default value of a leaf
func isDefaultValue(schemaEntry *yang.Entry, u *gnmi.Update) bool {
	if defval, singleDefVal := schemaEntry.SingleDefaultValue(); singleDefVal {
		return u.Val.GetStringVal() == defval
	}
	return false
}

// aggregates the proivided paths by extratcting relevant paths
func aggregateCommonPaths(p []*gnmi.Path) []*gnmi.Path {
	// we add the first path straight away to the list, therrefore make sure we have at least 1 entry in the list
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
			result = append(result, b.Elem[i])
		} else {
			break
		}
	}
	return result
}

// checks first for pointer equality or further for the Name, Key and Value based equality of two gnmi.PathElement structs
func pathElemIsEqual(a, b *gnmi.PathElem) bool {
	// the two variables are pointer equal
	if a == b {
		return true
	}

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

// returns a pointer to the srlygot root schema for later processing
func getRootSchema() *yang.Entry {
	schema, err := ygotsrl.Schema()
	if err != nil {
		panic(err)
	}

	deviceSchema := schema.RootSchema()
	return deviceSchema
}
