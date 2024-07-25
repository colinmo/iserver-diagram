package c4puml

import (
	"bytes"
	"fmt"
)

type Container struct {
	Model       string
	Alias       string
	Name        string
	TOGAF       string
	Tags        string
	Sprite      string
	Description string
	External    bool
}

type Boundary struct {
	Model      string
	Alias      string
	Name       string
	TOGAF      string
	Containers []Container
	Boundaries []Boundary
}

type Relationship struct {
	From       Container
	To         Container
	Label      string
	Technology string
	Direction  string
}

type Layout struct {
	From      Container
	To        Container
	Direction string
}

type Chart struct {
	Boundaries    []Boundary
	Containers    []Container
	Relationships []Relationship
	Layouts       []Layout
}

func relationshipAsString(x Relationship) string {
	toReturn := new(bytes.Buffer)

	if len(x.Direction) > 0 {
		x.Direction = fmt.Sprintf("_%s", x.Direction)
	}
	toReturn.WriteString(fmt.Sprintf("Rel%s(%s,%s,\"%s\",\"%s\")\n", x.Direction, x.From.Alias, x.To.Alias, x.Label, x.Technology))
	return toReturn.String()
}
func layoutsAsString(x Layout) string {
	toReturn := new(bytes.Buffer)

	toReturn.WriteString(fmt.Sprintf("Lay_%s(%s,%s)\n", x.Direction, x.From.Alias, x.To.Alias))
	return toReturn.String()
}

func boundaryAsString(b Boundary) string {
	toReturn := new(bytes.Buffer)
	toReturn.WriteString(fmt.Sprintf("System_Boundary(%s,\"%s\",$tags=\"%s\") { \n", b.Alias, b.Name, b.TOGAF))
	for _, b2 := range b.Boundaries {
		toReturn.WriteString(boundaryAsString(b2))
	}
	for _, c := range b.Containers {
		toReturn.WriteString(containerAsString(c))
	}
	return toReturn.String() + "}\n"
}

func containerAsString(c Container) string {
	return fmt.Sprintf(
		"System_Boundary(%s,\"%s\",$tags=\"%s\",$descr=\"%s\")\n",
		c.Alias,
		c.Name,
		c.TOGAF,
		c.Description,
	)
}

func NewChart() Chart {
	return Chart{}
}

func (c *Chart) Draw() string {
	toReturn := new(bytes.Buffer)
	toReturn.WriteString("@startuml Solution Context\n!include https://raw.githubusercontent.com/colinmo/iserver-diagram/main/togaf/togaf-full.puml\n")
	for _, ob := range c.Boundaries {
		toReturn.WriteString(boundaryAsString(ob))
	}
	for _, ob := range c.Containers {
		toReturn.WriteString(containerAsString(ob))
	}
	for _, ob := range c.Relationships {
		toReturn.WriteString(relationshipAsString(ob))
	}
	for _, ob := range c.Layouts {
		toReturn.WriteString(layoutsAsString(ob))
	}
	toReturn.WriteString("@enduml")
	return toReturn.String()
}
