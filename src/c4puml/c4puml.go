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

	switch b.Model {
	case "Enterprise":
		toReturn.WriteString("Enterprise_Boundary(")
		b.TOGAF = ""
	case "System":
		toReturn.WriteString("System_Boundary(")
		b.TOGAF = ""
	default:
		toReturn.WriteString("Boundary(")
	}
	toReturn.WriteString(fmt.Sprintf("%s,\"%s\",\"%s\"", b.Alias, b.Name, b.TOGAF))
	if len(b.TOGAF) > 0 {
		toReturn.WriteString(
			fmt.Sprintf(
				",$type=\"%s\"",
				b.TOGAF))
	}
	toReturn.WriteString(") {\n")
	for _, b2 := range b.Boundaries {
		toReturn.WriteString(boundaryAsString(b2))
	}
	for _, c := range b.Containers {
		toReturn.WriteString(containerAsString(c))
	}
	return toReturn.String() + "}\n"
}

func containerAsString(c Container) string {
	toReturn := new(bytes.Buffer)

	suffix := ""
	if c.External {
		suffix = "_Ext"
	}
	switch c.Model {
	case "ContainerDb":
		toReturn.WriteString(
			fmt.Sprintf(
				"ContainerDb%s(%s,\"%s\",\"\",\"%s\"",
				suffix,
				c.Alias,
				c.Name,
				c.Description))
	case "Person":
		toReturn.WriteString(
			fmt.Sprintf(
				"Person%s(%s,\"%s\",\"\",\"\"",
				suffix,
				c.Alias,
				c.Name,
			),
		)
	case "System":
		toReturn.WriteString(
			fmt.Sprintf(
				"System%s(%s,\"%s\",\"%s\",\"\"",
				suffix,
				c.Alias,
				c.Name,
				c.Description))
	}
	if len(c.TOGAF) > 0 {
		toReturn.WriteString(
			fmt.Sprintf(
				",$type=\"%s\"",
				c.TOGAF))
	}
	if len(c.Sprite) > 0 {
		toReturn.WriteString(
			fmt.Sprintf(
				",$sprite=\"%s\"",
				c.Sprite))
	}
	if len(c.Tags) > 0 {
		toReturn.WriteString(
			fmt.Sprintf(
				",$tags=\"%s\"",
				c.Tags))
	}
	toReturn.WriteString(")\n")
	return toReturn.String()
}

func NewChart() Chart {
	return Chart{}
}

func (c *Chart) Draw() string {
	toReturn := new(bytes.Buffer)
	toReturn.WriteString("@startuml Solution Context\n!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml\n!define DEVICONS https://raw.githubusercontent.com/tupadr3/plantuml-icon-font-sprites/master/devicons\n!define FONTAWESOME https://raw.githubusercontent.com/tupadr3/plantuml-icon-font-sprites/master/font-awesome-5\n!include DEVICONS/angular.puml\n!include DEVICONS/java.puml\n!include DEVICONS/msql_server.puml\n!include FONTAWESOME/users.puml\nSetDefaultLegendEntries(\"\")\nLAYOUT_WITH_LEGEND()\n")
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
