package c4puml

import (
	"log"
	"testing"
)

func TestDraw(t *testing.T) {
	d := NewChart()
	clnt := Container{
		Model:    "Person",
		Alias:    "clnt",
		Name:     "Client",
		TOGAF:    "role",
		Sprite:   "users",
		External: true,
	}
	pmgr := Container{
		Model: "Person",
		Alias: "pmgr",
		Name:  "Program Manager",
		TOGAF: "role",
	}
	pcch := Container{
		Model: "Person",
		Alias: "pcch",
		Name:  "Program Coach",
		TOGAF: "role",
	}
	tctla := Container{
		Model: "Person",
		Alias: "tctla",
		Name:  "Authorised TCTL staff",
		TOGAF: "role",
	}
	azdw := Container{
		Model:       "ContainerDb",
		Alias:       "AzureDW",
		Name:        "Database",
		Description: "Exported data storage vault",
	}
	ping := Container{
		Model:       "System",
		Alias:       "PingFederate",
		Name:        "Ping Federate",
		TOGAF:       "PTC",
		Description: "Authentication and authorisation",
	}
	gus := Boundary{
		Model: "System",
		Alias: "gus",
		Name:  "Griffith On-Prem Network",
		Containers: []Container{
			ping,
		},
	}
	redicase := Container{
		Model:    "System",
		Alias:    "rediCASE",
		Name:     "Case management system",
		TOGAF:    "PAC",
		Sprite:   "",
		External: true,
	}
	rda := Boundary{
		Model: "Enterprise",
		Alias: "rda",
		Name:  "Redbourne Azure (AE)",
		TOGAF: "LOC",
		Containers: []Container{
			redicase,
		},
		Boundaries: []Boundary{},
	}
	ga := Boundary{
		Model: "Enterprise",
		Alias: "ga",
		Name:  "Griffith Azure Tenancy",
		TOGAF: "",
		Containers: []Container{
			azdw,
		},
		Boundaries: []Boundary{},
	}
	gu := Boundary{
		Model: "Enterprise",
		Alias: "gu",
		Name:  "Griffith University",
		TOGAF: "",
		Containers: []Container{
			pmgr,
			pcch,
			tctla,
		},
		Boundaries: []Boundary{
			ga,
			gus,
		},
	}
	d.Boundaries = []Boundary{
		{
			Model: "",
			Alias: "casem",
			Name:  "Case Management",
			TOGAF: "LAC",
			Containers: []Container{
				clnt,
			},
			Boundaries: []Boundary{
				gu,
				rda,
			},
		},
	}
	d.Relationships = []Relationship{
		{From: pmgr, To: ping, Label: "Uses", Technology: "https"},
		{From: pcch, To: ping, Label: "Uses", Technology: "https"},
	}
	d.Layouts = []Layout{
		{From: ping, To: redicase, Direction: "Right"},
		{From: azdw, To: ping, Direction: "Up"},
	}
	x := d.Draw()
	log.Print("\n" + x)
}
