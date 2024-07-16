package main

type TogafIcon struct {
	Color string
	C4    string
}

var TogafIcons = map[string]TogafIcon{
	"pac": {
		Color: "#65b5f6",
		C4:    "System(%s,\"%s\",\"\",\"\",$tags=\"pac\",$type=\"Physical application component\") %s\n",
	},
	"lac": {
		Color: "#65b5f6",
		C4:    "System(%s,\"%s\",\"\",\"\",$tags=\"lac\",$type=\"Logical application component\") %s\n",
	},
	"ptc": {
		Color: "#02a89d",
		C4:    "System(%s,\"%s\",\"\",\"\",$tags=\"ptc\",$type=\"Physical technology component\") %s\n",
	},
	"act": {
		Color: "#00695c",
		C4:    "Person(%s,\"%s\",\"\",$tags=\"pdc\",$type=\"Actor\") %s\n",
	},
	"loc": {
		Color: "#623f36",
		C4:    "Enterprise_Boundary(%s,\"%s\",\"location\") { %s }\n",
	},
}
