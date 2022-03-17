package main

import (
	"fmt"
	"sort"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"
)

var rootCE *configElement

type MyPathElem struct {
	*gnmi.PathElem
}

func (pe *MyPathElem) hashCode() string {

	result := "["
	sep := ""

	keys := make([]string, 0, len(pe.Key))

	for k := range pe.Key {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result += fmt.Sprintf("%s: %s%s", k, pe.Key[k], sep)
		sep = ","
	}
	result += "]"

	return fmt.Sprintf(pe.Name + result)
}

type configElement struct {
	*PathAndSchema
	children map[string]*configElement
	parent   *configElement
	name     string
	keys     map[string]string
}

func (ce *configElement) getHierarchicalOutput(level int, indentor string) string {
	result := ""
	indent := ""
	for i := 0; i <= level; i++ {
		indent += indentor
	}
	result += fmt.Sprintf("%s %s %s {hasNonKeyChilds: %s, IsLeaf: %s, IsKey: %s, hasOnlyKeyChilds: %s, hasLeafChilds: %s, hasNonKeyLeafChilds: %s}\n", indent, ce.name, ce.getKeysAsString(),
		bool2string(ce.hasNonKeyChilds()),
		bool2string(ce.isLeaf()),
		bool2string(ce.isKey()),
		bool2string(ce.hasOnlyKeyChilds()),
		bool2string(ce.hasLeafChilds()),
		bool2string(ce.hasNonKeyLeafChilds()),
	)
	for _, cce := range ce.children {
		result += cce.getHierarchicalOutput(level+1, indentor)
	}
	return result
}

func (ce *configElement) isKey() bool {
	result := false
	if ce.parent == nil {
		// on root node return false
		return false
	}
	for k, _ := range ce.parent.keys {
		if ce.name == k {
			result = true
		}
	}
	return result
}

func (ce *configElement) getChildCount() int {
	return len(ce.children)
}

func (ce *configElement) isLeaf() bool {
	return len(ce.children) <= 0
}

func (ce *configElement) hasNonKeyChilds() bool {
	result := false
	if ce.getChildCount() == 0 {
		return !ce.isKey()
	} else {
		for _, v := range ce.children {
			result = result || v.hasNonKeyChilds()
		}
	}

	return result
}

func (ce *configElement) getRootPaths() []*gnmi.Path {
	result := []*gnmi.Path{}

	// skip leaf elements, they should not appear in the result
	if ce.isLeaf() {
		return result
	}
	if ce.hasNonKeyLeafChilds() {
		result = append(result, ce.path)
	} else if ce.hasOnlyKeyChilds() {
		result = append(result, ce.path)
	} else {
		for _, v := range ce.children {
			result = append(result, v.getRootPaths()...)
		}
	}

	return result
}
func (ce *configElement) hasNonKeyLeafChilds() bool {
	for _, c := range ce.children {
		if c.isLeaf() && !c.isKey() {
			return true
		}
	}
	return false
}
func (ce *configElement) hasLeafChilds() bool {
	for _, c := range ce.children {
		if c.isLeaf() {
			return true
		}
	}
	return false

}

func (ce *configElement) hasOnlyKeyChilds() bool {
	for _, c := range ce.children {
		if !c.isKey() {
			return false
		}
	}
	return true && len(ce.children) > 0
}

func (ce *configElement) getKeysAsString() string {
	result := ""
	if len(ce.keys) > 0 {
		result = "["
		sep := ""
		for k, v := range ce.keys {
			result += fmt.Sprintf("%s%s: %s", sep, k, v)
			sep = ", "
		}
		result += "]"
	}
	return result
}

func (ce *configElement) getHierarchicalName() string {
	parentname := ""
	if ce.parent != nil {
		parentname = ce.parent.getHierarchicalName() + "->"
	}

	return parentname + ce.name + ce.getKeysAsString()
}

func (ce *configElement) String() string {
	return fmt.Sprintf("%s Level: %d", ce.getHierarchicalName(), ce.getLevel())
}

func (ce *configElement) getLevel() int {
	if ce.parent == nil {
		return -1
	}
	parentlevel := ce.parent.getLevel()
	return parentlevel + 1
}

func NewConfigElement(name string, parent *configElement, keys map[string]string) *configElement {
	return &configElement{name: name, children: map[string]*configElement{}, parent: parent, PathAndSchema: nil, keys: keys}
}

func (ce *configElement) add(pas *PathAndSchema, index int) {
	if len(pas.path.Elem)-1 == ce.getLevel() {
		return
	}
	name := pas.path.Elem[index].Name
	elemPath := &MyPathElem{pas.path.Elem[index]}

	val, exists := ce.children[elemPath.hashCode()]

	if !exists {
		val = NewConfigElement(name, ce, elemPath.Key)
		val.PathAndSchema = getPathAndSchemaEntry(getRootSchemafromIntermediate(pas.schema), &gnmi.Path{Elem: pas.path.Elem[0 : index+1]})
		ce.children[elemPath.hashCode()] = val
	}
	fmt.Println(val.String())
	val.add(pas, index+1)
}

func getRootSchemafromIntermediate(se *yang.Entry) *yang.Entry {
	tmp := se
	for tmp.Parent != nil {
		tmp = tmp.Parent
	}
	return tmp
}

func getRootConfigElement() *configElement {
	return &configElement{
		children: map[string]*configElement{},
		parent:   nil,
		name:     "Device",
		PathAndSchema: &PathAndSchema{
			path:   &gnmi.Path{},
			schema: getRootSchema(),
		},
	}
}

func runhierarchical(pas *PathAndSchema) {
	rootCE.add(pas, 0)
}

func init() {
	rootCE = getRootConfigElement()
}
