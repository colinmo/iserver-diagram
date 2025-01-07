package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	azure "vonexplaino.com/m/v2/vondiagram/azure"
)

func TestCreateHERMHTML(t *testing.T) {
	objs := []azure.ObjectStruct{
		{
			ObjectID:        "1",
			Name:            "PAC1",
			ObjectTypeId:    "6fb624e4-b642-ea11-a601-28187852aafd",
			ObjectType:      azure.ObjectTypeStruct{Name: "Physical Application Component"},
			Attributevalues: []azure.AttributeTypeStruct{},
		},
		{
			ObjectID:        "2",
			Name:            "PAC2",
			ObjectTypeId:    "6fb624e4-b642-ea11-a601-28187852aafd",
			ObjectType:      azure.ObjectTypeStruct{Name: "Physical Application Component"},
			Attributevalues: []azure.AttributeTypeStruct{},
		},
	}
	lnks := []azure.MinRelationship{
		{LeadObjectID: "1", MemberObjectID: "2"},
	}
	bob := createHERMHTML(objs, lnks)
	assert.Contains(t, bob, `<html>`)
	assert.Contains(t, bob, "const nodes = [\n            { id: \"1\", group: 1, name: \"PAC1\", type: \"Physical Application Component\" },{ id: \"2\", group: 1, name: \"PAC2\", type: \"Physical Application Component\" },\n        ];\n        const links = [\n            { source: \"1\", target: \"2\", value: \"9\" },\n            \n        ];")
}
