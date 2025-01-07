package main

import (
	"bytes"
	"os"
	"text/template"

	azure "vonexplaino.com/m/v2/vondiagram/azure"
)

/**
** Create an HTML report on the HERM relations in a domain
** Uses D3 to create a force directed graph of relationships
**/

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
	type graphicStruct struct {
		OBJS []azure.ObjectStruct
		LNKS []azure.MinRelationship
	}
	var tmplFile = "force-graph.html"
	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, graphicStruct{
		OBJS: objs,
		LNKS: lnks,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
