package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"

	azure "vonexplaino.com/m/v2/vondiagram/azure"
)

/**
** Create an HTML report on the HERM relations in a domain
** Uses D3 to create a force directed graph of relationships
**/

//go:embed force-graph.html
var tmplFile string

func CreateHERM() {
	// Download iServer data
	objects := az.GetDomainObjectsForHERM()
	// Get relationships
	objects, relations := az.GetRelatedHERMObjects(objects)
	// Convert into the D3 expected format
	// Save to HTML
	os.WriteFile("/Users/s457972/Dropbox/swap/golang/von-iserver-diagram/src/force-graph-out.html", []byte(createHERMHTML(objects, relations)), 0644)
}

func createHERMHTML(objs []azure.ObjectStruct, lnks []azure.MinRelationship) string {
	// Setup variables
	type graphicStruct struct {
		OBJS    []azure.ObjectStruct
		LNKS    []azure.MinRelationship
		PACRELS map[string]struct {
			Name      string
			Relations map[string]string
		}
	}
	gs := graphicStruct{
		OBJS: objs,
		LNKS: lnks,
		PACRELS: map[string]struct {
			Name      string
			Relations map[string]string
		}{},
	}
	indexedObjs := map[string]azure.ObjectStruct{}
	// Create table by getting all PACs and then getting any linked items
	for _, ob := range objs {
		indexedObjs[ob.ObjectID] = ob
		if ob.ObjectType.Name == "Physical Application Component" {
			gs.PACRELS[ob.ObjectID] = struct {
				Name      string
				Relations map[string]string
			}{Name: ob.Name,
				Relations: map[string]string{
					"Physical Technology Component": "<br>",
					"Capability":                    "<br>",
					"Physical Data Component":       "<br>",
					"Logical Application Component": "<br>",
				},
			}
		}
	}
	for _, ln := range lnks {
		if _, ok := gs.PACRELS[ln.LeadObjectID]; ok {
			gs.PACRELS[ln.LeadObjectID].Relations[indexedObjs[ln.MemberObjectID].ObjectType.Name] = fmt.Sprintf("%s%s<br>", gs.PACRELS[ln.LeadObjectID].Relations[indexedObjs[ln.MemberObjectID].ObjectType.Name], indexedObjs[ln.MemberObjectID].Name)
		}
		if _, ok := gs.PACRELS[ln.MemberObjectID]; ok {
			gs.PACRELS[ln.MemberObjectID].Relations[indexedObjs[ln.LeadObjectID].ObjectType.Name] = fmt.Sprintf("%s%s<br>", gs.PACRELS[ln.MemberObjectID].Relations[indexedObjs[ln.LeadObjectID].ObjectType.Name], indexedObjs[ln.LeadObjectID].Name)
		}
	}

	tmpl, err := template.New(tmplFile).Parse(tmplFile)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, gs)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
