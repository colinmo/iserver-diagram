package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	azure "vonexplaino.com/m/v2/vondiagram/azure"
	"vonexplaino.com/m/v2/vondiagram/mywidge"
)

var status binding.String
var messages binding.String
var az azure.AzureAuth
var centreContent *fyne.Container
var myApp fyne.App
var windows map[string]fyne.Window

type objectStruct struct {
	alias    string
	name     string
	otype    string
	children map[string]objectStruct
}

type relationshipStruct struct {
	leftAlias        string
	rightAlias       string
	relationshipName string
}

func main() {
	// Basic window setup
	windows = make(map[string]fyne.Window)
	myApp = app.NewWithID("com.vonexplaino.voniserverdiagram")
	status = binding.NewString()
	messages = binding.NewString()
	dept := widget.NewSelect([]string{}, func(change string) {})
	if myApp.Preferences().StringWithFallback("Department", "nope") == "nope" {
		UpdateStatus("Offline")
		UpdateMessage("Update settings")
	} else {
		// In background, start logging in
		go func() {
			UpdateStatus("Loading")
			UpdateMessage("Waiting")
			az.StartAzure()
			UpdateStatus("Live")
			UpdateMessage("Ready")
			keys := getMapStringKeys(az.GetChoicesForName("GU::Domain"))
			sort.Strings(keys)
			dept.Options = keys
			dept.SetSelected(myApp.Preferences().StringWithFallback("Department", "Unknown"))
			dept.Refresh()
		}()
	}
	mainWindow := myApp.NewWindow("von iServer")
	mainWindow.Resize(fyne.NewSize(600, 600))
	mainWindow.SetCloseIntercept(func() {
		if len(windows) == 0 {
			mainWindow.Close()
		}
	})
	bottom := container.New(
		layout.NewHBoxLayout(),
		widget.NewLabelWithData(status),
		layout.NewSpacer(),
		widget.NewLabelWithData(messages),
	)
	searchEntry := binding.NewString()
	centreContent = container.NewStack()
	// Settings
	pms := widget.NewMultiLineEntry()
	pms.SetText(myApp.Preferences().StringWithFallback("ProductManagers", "[]"))
	pms.SetMinRowsVisible(7)
	savepath := widget.NewEntry()
	savepath.SetText(myApp.Preferences().StringWithFallback("SavePath", ""))
	searchButton := widget.NewButton(
		"Go",
		func() {
			if x, _ := status.Get(); x == "Live" {
				UpdateMessage("Searching...")
				text, _ := searchEntry.Get()
				go func() {
					az.FindMeThen(text, ListAndSelectAThing, &mainWindow)
					UpdateMessage("Ready")
				}()
			} else {
				UpdateMessage("Not ready")
			}
		},
	)
	searchBox := newEnterEntryWithData(searchEntry, searchButton)
	tabs := container.NewAppTabs(
		container.NewTabItem(
			"Diagrams",
			container.NewBorder(
				container.NewBorder(
					widget.NewToolbar(
						widget.NewToolbarAction(
							resourcePacPng,
							func() {
								windowTitle := "New Physical Application Component"
								var lookupWindow fyne.Window
								var x bool
								if lookupWindow, x = windows[windowTitle]; !x {
									addWindowFor(windowTitle, 650, 850)
									lookupWindow = windows[windowTitle]
								}
								def := newPACTemplate(PacFields())
								UpdateMessage("Loading")
								lookupWindow.Show()
								lookupWindow.SetContent(makeLookupWindow(widget.NewLabel("Loading...")))
								windows[windowTitle] = lookupWindow
								ListRelationsToSelect(def, []azure.RelationStruct{}, &lookupWindow)
							},
						),
						widget.NewToolbarAction(
							resourcePtcPng,
							func() {
								windowTitle := "New Physical Technology Component"
								var lookupWindow fyne.Window
								var x bool
								if lookupWindow, x = windows[windowTitle]; !x {
									addWindowFor(windowTitle, 650, 850)
									lookupWindow = windows[windowTitle]
								}
								def := newPTCTemplate(PtcFields())
								UpdateMessage("Loading")
								lookupWindow.Show()
								lookupWindow.SetContent(makeLookupWindow(widget.NewLabel("Loading...")))
								windows[windowTitle] = lookupWindow
								ListRelationsToSelect(def, []azure.RelationStruct{}, &lookupWindow)
							},
						),
					),
					nil,
					widget.NewLabel("Looking for"),
					searchButton,
					searchBox,
				),
				nil,
				nil,
				nil,
				centreContent,
			)),
		container.NewTabItem(
			"Audit",
			container.NewGridWrap(
				fyne.Size{Width: 160, Height: 40},
				widget.NewButton("Product Managers", func() {
					UpdateMessage("Loading")
					az.GetProductManagersThen(myApp.Preferences().StringWithFallback("Department", ""), ShowManagersList, &mainWindow)
					UpdateMessage("Ready")
				}),
				widget.NewButton("Domain audit", func() {
					UpdateMessage("Loading")
					thewindow := addWindowFor("Apps by PM", 300, 500)
					thewindow.SetContent(widget.NewLabel("Loading..."))
					thewindow.Show()
					az.GetDomainThen(myApp.Preferences().StringWithFallback("Department", ""), ShowDomainTree, thewindow)
					UpdateMessage("Ready")
				}),
				widget.NewButton("Excel Audit", func() {
					UpdateMessage("Running")
					az.CreateProductManagerOverviewReport(myApp.Preferences().StringWithFallback("Department", "nope"), getSavePath())
					UpdateMessage("Complete")
				}),
			)),
		container.NewTabItem(
			"Settings",
			container.NewVScroll(
				container.NewVBox(
					widget.NewForm(
						widget.NewFormItem("Domains", dept),
						widget.NewFormItem("Product Managers", pms),
						widget.NewFormItem("Save path", savepath),
					),
					widget.NewButton("Save", func() {
						myApp.Preferences().SetString("Department", dept.Selected)
						myApp.Preferences().SetString("ProductManagers", PrettyJSONString(pms.Text))
						myApp.Preferences().SetString("SavePath", savepath.Text)
					})),
			)),
	)
	mainWindow.SetContent(
		container.NewBorder(
			nil,
			bottom,
			nil,
			nil,
			tabs,
		),
	)

	// Display
	mainWindow.Show()
	myApp.Run()
	tidyUp()
}

func newPACTemplate(template modelFields) azure.IServerObjectStruct {
	newObject := azure.IServerObjectStruct{
		Name:     "",
		ObjectId: "",
		AttributeValues: []azure.AttributeValue{
			{StringValue: "", AttributeName: "Owner"},
		},
		ObjectType: struct {
			Name string `json:"Name"`
			Id   string `json:"ObjectTypeId"`
		}{Name: "Physical Application Component"},
	}
	for name := range template.selectValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	for name := range template.radioValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	for name := range template.checkValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	return newObject
}

func newPTCTemplate(template modelFields) azure.IServerObjectStruct {
	newObject := azure.IServerObjectStruct{
		Name:     "",
		ObjectId: "",
		AttributeValues: []azure.AttributeValue{
			{StringValue: "", AttributeName: "Owner"},
		},
		ObjectType: struct {
			Name string `json:"Name"`
			Id   string `json:"ObjectTypeId"`
		}{Name: "Physical Technology Component"},
	}
	for name := range template.selectValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	for name := range template.radioValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	for name := range template.checkValues {
		azure.ValidChoices[name] = az.GetChoicesForName(name)
		newObject.AttributeValues = append(
			newObject.AttributeValues,
			azure.AttributeValue{
				AttributeName: name,
			},
		)
	}
	return newObject
}
func tidyUp() {
	fmt.Println("Exited")
}

func UpdateStatus(newStatus string) {
	status.Set(newStatus)
}

func UpdateMessage(newMessage string) {
	messages.Set(newMessage)
}

func ListAndSelectAThing(things []azure.FindStruct, thenWindow *fyne.Window) {
	display := widget.NewList(
		func() int { return len(things) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewButtonWithIcon("Load", theme.ComputerIcon(), func() {}),
				widget.NewLabel("template"),
			)
		},
		func(id int, item fyne.CanvasObject) {
			me := item.(*fyne.Container).Objects[0].(*widget.Button)
			switch things[id].Type.Name {
			case "Physical Technology Component":
				me.SetIcon(resourcePtcPng)
				me.SetText("PTC")
			case "Physical Application Component":
				me.SetIcon(resourcePacPng)
				me.SetText("PAC")
			case "Logical Application Component":
				me.SetIcon(resourceLacPng)
				me.SetText("LAC")
			default:
				fmt.Printf("Dunno: %s\n", things[id].Type.Name)
			}
			me.OnTapped = func() {
				if me.Text == "PAC" {
					windowTitle := fmt.Sprintf("Details for %s", things[id].Name)
					var lookupWindow fyne.Window
					var x bool
					if lookupWindow, x = windows[windowTitle]; !x {
						addWindowFor(windowTitle, 650, 920)
						lookupWindow = windows[windowTitle]
					}
					UpdateMessage("Loading")
					lookupWindow.Show()

					lookupWindow.SetContent(makeLookupWindow(widget.NewLabel("Loading...")))
					windows[windowTitle] = lookupWindow
					az.FindRelationsThen(things[id].ObjectId, "PAC", ListRelationsToSelect, &lookupWindow)
					UpdateMessage("Ready")
				} else {
					dialog.ShowInformation("Nope", "Physical Application Containers only for now sorry", *thenWindow)
				}
			}
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(things[id].Name)
		},
	)
	ChangeMiddleContent(display, thenWindow)
}

type fieldsStruct struct {
	label       string
	fieldindex  string
	valuesIndex string
}

type sectionStruct struct {
	title  string
	fields map[int]map[int]fieldsStruct
}

type modelFields struct {
	stringValues map[string]*widget.Entry
	selectValues map[string]*widget.Select
	radioValues  map[string]*widget.RadioGroup
	checkValues  map[string]*widget.CheckGroup
	dateValues   map[string]*widget.Entry
	sections     map[int]sectionStruct
}

func ListRelationsToSelect(
	basics azure.IServerObjectStruct,
	things []azure.RelationStruct,
	thenWindow *fyne.Window,
) {

	var allFields modelFields
	switch basics.ObjectType.Name {
	case "Physical Application Component":
		allFields = PacFields()
	case "Physical Technology Component":
		allFields = PtcFields()
	}
	allFields.stringValues["Description"].Wrapping = fyne.TextWrapWord
	allFields.stringValues["Description"].SetMinRowsVisible(5)
	json.Unmarshal([]byte(myApp.Preferences().StringWithFallback("ProductManagers", "[]")), &allFields.selectValues["Owner"].Options)
	allFields.stringValues["Title"].SetText(basics.Name)
	selectedRelations := map[string]azure.RelationStruct{}
	for i := range allFields.dateValues {
		allFields.dateValues[i].Validator = dateValidator
	}
	isString := func(str string) bool { _, x := allFields.stringValues[str]; return x }
	isSelect := func(str string) bool { _, x := allFields.selectValues[str]; return x }
	isRadio := func(str string) bool { _, x := allFields.radioValues[str]; return x }
	isCheck := func(str string) bool { _, x := allFields.checkValues[str]; return x }
	isDate := func(str string) bool { _, x := allFields.dateValues[str]; return x }
	for _, x := range basics.AttributeValues {
		switch {
		case isString(x.AttributeName):
			allFields.stringValues[x.AttributeName].SetText(x.StringValue)
		case isSelect(x.AttributeName):
			if x.AttributeName == "Build" {
				x.StringValue = strings.Split(x.StringValue, " ")[0]
			}
			if x.AttributeName != "Owner" && x.AttributeName != "GU::Managed outside of DS" {
				azure.ValidChoices[x.AttributeName] = az.GetChoicesForName(x.AttributeName)
				keys := getMapStringKeys(azure.ValidChoices[x.AttributeName])
				sort.Strings(keys)
				allFields.selectValues[x.AttributeName] = widget.NewSelect(
					keys,
					func(bob string) {},
				)
			}
			allFields.selectValues[x.AttributeName].Selected = x.StringValue
		case isRadio(x.AttributeName):
			azure.ValidChoices[x.AttributeName] = az.GetChoicesForName(x.AttributeName)
			keys := getMapStringKeys(azure.ValidChoices[x.AttributeName])
			sort.Strings(keys)
			allFields.radioValues[x.AttributeName] = widget.NewRadioGroup(
				keys,
				func(bob string) {},
			)
			allFields.radioValues[x.AttributeName].Selected = x.StringValue
		case isCheck(x.AttributeName):
			azure.ValidChoices[x.AttributeName] = az.GetChoicesForName(x.AttributeName)
			keys := getMapStringKeys(azure.ValidChoices[x.AttributeName])
			sort.Strings(keys)
			allFields.checkValues[x.AttributeName] = widget.NewCheckGroup(
				keys,
				func(bob []string) {},
			)
			allFields.checkValues[x.AttributeName].Horizontal = true
			allFields.checkValues[x.AttributeName].Selected = []string{}
			for _, elem := range strings.Split(x.StringValue, ",") {
				allFields.checkValues[x.AttributeName].Selected = append(
					allFields.checkValues[x.AttributeName].Selected,
					strings.Trim(elem, " "),
				)
			}
		case isDate(x.AttributeName):
			allFields.dateValues[x.AttributeName].SetText(strings.Replace(x.StringValue, "T00:00:00Z", "", 1))
		default:
			message := fmt.Sprintf("Don't know who %s is\n", x.AttributeName)
			fyne.LogError(message, fmt.Errorf(message))
		}
	}
	knownKids := map[widget.TreeNodeID][]widget.TreeNodeID{}
	knownBits := map[widget.TreeNodeID]azure.RelationStruct{}
	for _, x := range things {
		knownKids[""] = append(knownKids[""], x.RelationshipId)
		knownBits[x.RelationshipId] = x
	}
	relationshipWindow := createRelationshipWindow(
		basics,
		selectedRelations,
		knownKids,
		knownBits,
		thenWindow)
	display := container.NewBorder(
		widget.NewToolbar(
			widget.NewToolbarAction(
				theme.ComputerIcon(),
				func() {
					openbrowser(fmt.Sprintf("https://griffith.iserver365.com/object/%s/details", basics.ObjectId))
				},
			),
			widget.NewToolbarAction(
				theme.DocumentSaveIcon(),
				func() {
					var d dialog.Dialog
					d = dialog.NewConfirm("Save", "Are you sure you want to commit these changes?", func(ok bool) {
						if ok {
							stringValuesAsString := map[string]string{}
							selectValuesAsString := map[string]string{}
							dateValuesAsString := map[string]string{}
							for i, x := range allFields.stringValues {
								stringValuesAsString[i] = x.Text
							}
							for i, x := range allFields.selectValues {
								if x.Selected != "" {
									selectValuesAsString[i] = x.Selected
								}
							}
							for i, x := range allFields.radioValues {
								if x.Selected != "" {
									stringValuesAsString[i] = x.Selected
								}
							}
							for i, x := range allFields.checkValues {
								if len(x.Selected) > 0 {
									selectValuesAsString[i] = strings.Join(x.Selected, ",")
								}
							}
							for i, x := range allFields.dateValues {
								dateValuesAsString[i] = x.Text
							}
							title := "Save Succesful"
							_, message, id := az.SaveObjectFields(
								basics.ObjectId,
								basics.ObjectType.Name,
								stringValuesAsString,
								selectValuesAsString,
								dateValuesAsString,
							)
							basics.ObjectId = id
							d.Hide()
							d2 := dialog.NewInformation(title, message, *thenWindow)
							d2.Show()
						}
					}, *thenWindow)
					d.Show()
				},
			),
		),
		nil,
		nil,
		nil,
		container.NewVScroll(makeEditPage(allFields, relationshipWindow, thenWindow)))
	(*thenWindow).SetContent(makeLookupWindow(display))
}

func makeEditPage(allFields modelFields, relationshipWindow *fyne.Container, thisWindow *fyne.Window) *container.AppTabs {
	box := container.NewAppTabs()
	for i := 0; i < len(allFields.sections); i++ {
		baseform := container.NewVBox()
		sec := allFields.sections[i]
		for k := 0; k < len(sec.fields); k++ {
			row := sec.fields[k]
			rowform := container.NewGridWithColumns(len(row))
			for j := 0; j < len(row); j++ {
				fld := row[j]
				label := widget.NewLabelWithStyle(fld.label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				switch fld.fieldindex {
				case "string":
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, allFields.stringValues[fld.valuesIndex]))
				case "select":
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, allFields.selectValues[fld.valuesIndex]))
				case "radio":
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, allFields.radioValues[fld.valuesIndex]))
				case "check":
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, allFields.checkValues[fld.valuesIndex]))
				case "date":
					allFields.dateValues[fld.valuesIndex] = mywidge.CalendarEntry(allFields.dateValues[fld.valuesIndex].Text, *thisWindow)
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, allFields.dateValues[fld.valuesIndex]))
				default:
					rowform.Objects = append(rowform.Objects, container.NewBorder(label, nil, nil, nil, widget.NewLabel("Unknown "+fld.fieldindex)))
				}
			}
			baseform.Objects = append(baseform.Objects, rowform)
		}
		box.Append(container.NewTabItem(
			sec.title,
			baseform,
		))
	}
	box.Append(container.NewTabItem("Relationships", relationshipWindow))
	return box
}

func ChangeMiddleContent(me fyne.CanvasObject, theWindow *fyne.Window) {
	UpdateStatus("Refreshing...")
	centreContent.Objects = []fyne.CanvasObject{me}
	centreContent.Refresh()
	UpdateStatus("Live")
}

func ShowDomainTree(things map[string][]azure.IServerObjectStruct, thenWindow fyne.Window) {
	parents := getMapInterfaceKeys(things)
	sort.Strings(parents)
	tree := widget.NewTree(
		func(tni widget.TreeNodeID) []widget.TreeNodeID {
			if tni == "" {
				return append([]widget.TreeNodeID{}, parents...)
			}
			returning := []widget.TreeNodeID{}
			for _, x := range things[tni] {
				returning = append(returning, fmt.Sprintf("%s~%s", x.Name, x.ObjectId))
			}
			sort.Strings(returning)
			return returning
		},
		func(tni widget.TreeNodeID) bool {
			if tni == "" {
				return true
			}
			for i := range things {
				if i == tni {
					return true
				}
			}
			return false
		},
		func(b bool) fyne.CanvasObject {
			if b {
				return widget.NewLabel("Branch template")
			}
			return widget.NewButton("Leaf template", func() {})
		},
		func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
			if b {
				co.(*widget.Label).SetText(tni)
			} else {
				meps := strings.Split(tni, "~")
				co.(*widget.Button).SetText(meps[0])
				co.(*widget.Button).OnTapped = func() {
					windowTitle := fmt.Sprintf("Details for %s", meps[0])
					var lookupWindow fyne.Window
					var x bool
					if lookupWindow, x = windows[windowTitle]; !x {
						addWindowFor(windowTitle, 650, 850)
						lookupWindow = windows[windowTitle]
					}
					UpdateMessage("Loading")
					lookupWindow.Show()

					lookupWindow.SetContent(makeLookupWindow(widget.NewLabel("Loading...")))
					windows[windowTitle] = lookupWindow
					az.FindRelationsThen(meps[1], "GEN", ListRelationsToSelect, &lookupWindow)
					UpdateMessage("Ready")
				}
			}
		},
	)
	(thenWindow).SetContent(tree)
	(thenWindow).Show()
}

func makeLookupWindow(contents fyne.CanvasObject) fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		contents,
	)
}

func ShowManagersList(list map[string][]string, window *fyne.Window) {
	windowTitle := "Manager List"
	addWindowFor(windowTitle, 300, 500)
	keys := getMapKeys(list)
	pms := []string{}
	pmslist := myApp.Preferences().String("ProductManagers")
	json.Unmarshal([]byte(pmslist), &pms)
	listlist := widget.NewList(
		func() int { return len(list) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id int, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(keys[id])
		},
	)
	var selectedThing string
	listlist.OnSelected = func(id widget.ListItemID) {
		selectedThing = keys[id]
	}
	windows[windowTitle].SetContent(
		container.NewBorder(widget.NewForm(
			widget.NewFormItem("Change PM", widget.NewSelect(pms, func(s string) {
				fmt.Printf("Change the list to %s\n", s)
				fmt.Printf("SI %v\n", selectedThing)
				fmt.Printf("For all %v\n", list[selectedThing])
			})),
			widget.NewFormItem("", widget.NewButton("Update", func() {})),
		),

			nil,
			nil,
			nil,
			container.NewVScroll(listlist)))
	windows[windowTitle].Show()
}

func addWindowFor(title string, w, h float32) fyne.Window {
	if _, x := windows[title]; !x {
		showWindow := myApp.NewWindow(title)
		showWindow.Resize(fyne.NewSize(w, h))
		showWindow.SetOnClosed(func() {
			delete(windows, title)
		})
		windows[title] = showWindow
	}
	return windows[title]
}

func getMapKeys(me map[string][]string) []string {
	keys := make([]string, len(me))

	i := 0
	for k := range me {
		keys[i] = k
		i++
	}
	return keys
}

func getMapInterfaceKeys(me map[string][]azure.IServerObjectStruct) []string {
	keys := make([]string, len(me))

	i := 0
	for k := range me {
		keys[i] = k
		i++
	}
	return keys
}

func getMapStringKeys(me map[string]string) []string {
	toReturn := []string{}
	for i := range me {
		toReturn = append(toReturn, i)
	}
	return toReturn
}

func nameToToken(alreadyDrawn *map[string]string, name string) string {
	bob := regexp.MustCompile("[^a-zA-Z0-9]")
	attempt := bob.ReplaceAllString(name, "")
	ok := false
	for !ok {
		ok = true
		for _, x := range *alreadyDrawn {
			if x == attempt {
				attempt += "1"
				ok = false
				break
			}
		}
	}
	return attempt
}

func drawObject(object objectStruct) string {
	defaultObjects := map[string]string{
		"actor":                          "act",
		"application service":            "aps",
		"business service":               "bus",
		"capability":                     "cap",
		"constraint":                     "cnt",
		"data entity":                    "dte",
		"interface":                      "int",
		"organization unit":              "org",
		"physical application component": "pac",
		"physical data component":        "pdc",
		"physical technology component":  "ptc",
		"physical technology group":      "ptg",
		"principle":                      "prn",
		"process":                        "pro",
		"product":                        "prd",
		"requirement":                    "req",
		"risk":                           "rsk",
		"role":                           "rol",
		"technology service":             "tcs",
	}
	fmt.Printf("Mapping %s\n", object.otype)
	switch strings.ToLower(object.otype) {
	case "location":
		children := strings.Builder{}
		for _, x := range object.children {
			children.WriteString(drawObject(x))
		}
		return fmt.Sprintf(
			"System_Boundary(%s,\"%s\",$tags=\"loc\") {\n%s}\n",
			object.alias,
			object.name,
			children.String(),
		)
	case "logical application component":
		children := strings.Builder{}
		for _, x := range object.children {
			children.WriteString(drawObject(x))
		}
		return fmt.Sprintf(
			"System_Boundary(%s,\"%s\",$tags=\"loc\") {\n%s}\n",
			object.alias,
			object.name,
			children.String(),
		)
	default:
		return fmt.Sprintf(
			"System_Boundary(%s,\"%s\",$tags=\"%v\")\n",
			object.alias,
			object.name,
			defaultObjects[strings.ToLower(object.otype)],
		)
	}
}

func addToObjectStruct(allObjects *map[string]objectStruct, alias, name, ptype string) {
	bob := make(map[string]objectStruct, 0)
	(*allObjects)[alias] = objectStruct{
		alias,
		name,
		ptype,
		bob,
	}
}

func addRelationship(
	relationships *map[string]relationshipStruct,
	objects *map[string]objectStruct,
	leftAlias string,
	leftObject azure.FindStruct,
	rightAlias string,
	rightObject azure.FindStruct,
	relationshipId string,
	connection string,
) {
	objectsRef := (*objects)
	if leftObject.Type.Name == "Location" {
		objectsRef[leftAlias].children[rightObject.ObjectId] = (*objects)[rightAlias]
		delete(objectsRef, rightAlias)
	} else if rightObject.Type.Name == "Location" {
		objectsRef[rightAlias].children[leftObject.ObjectId] = (*objects)[leftAlias]
		delete(objectsRef, leftAlias)
	} else {
		(*relationships)[relationshipId] = relationshipStruct{
			leftAlias:        leftAlias,
			rightAlias:       rightAlias,
			relationshipName: connection,
		}
	}
}

func createRelationshipList(
	selectedRelations map[string]azure.RelationStruct,
	knownKids map[widget.TreeNodeID][]widget.TreeNodeID,
	knownBits map[widget.TreeNodeID]azure.RelationStruct) *widget.Tree {
	return widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			y, x := knownKids[id]
			if x {
				return y
			}
			return []widget.TreeNodeID{}
		},
		func(id widget.TreeNodeID) bool {
			_, here := knownKids[id]
			return here
		},
		func(branch bool) fyne.CanvasObject {
			return container.NewHBox(widget.NewCheck("Diag", func(value bool) {}))
		},
		func(id widget.TreeNodeID, branch bool, item fyne.CanvasObject) {
			checkbox := &(item.(*fyne.Container).Objects[0])
			(*checkbox).(*widget.Check).OnChanged = func(value bool) {
				if value {
					selectedRelations[knownBits[id].RelationshipId] = knownBits[id]
					_, here := knownKids[knownBits[id].RelationshipId]
					if !here {
						go func() {
							knownKids[knownBits[id].RelationshipId] = []widget.TreeNodeID{}
							rels := append(az.FindRelations(knownBits[id].LeadObjectId), az.FindRelations(knownBits[id].MemberObjectId)...)
							for _, x := range rels {
								_, here2 := knownBits[x.RelationshipId]
								if !here2 {
									knownKids[knownBits[id].RelationshipId] = append(knownKids[knownBits[id].RelationshipId], x.RelationshipId)
									knownBits[x.RelationshipId] = x
								}
							}
						}()
					}
				} else {
					delete(selectedRelations, knownBits[id].RelationshipId)
				}
			}
			(*checkbox).(*widget.Check).Text = (fmt.Sprintf(
				"%s %s %s",
				knownBits[id].LeadObject.Name,
				knownBits[id].RelationshipType.LeadToMemberDirection,
				knownBits[id].MemberObject.Name,
			))
			_, x := selectedRelations[knownBits[id].RelationshipId]
			(*checkbox).(*widget.Check).SetChecked(x)
			(*checkbox).Refresh()
		},
	)
}

func createRelationshipWindow(
	basics azure.IServerObjectStruct,
	selectedRelations map[string]azure.RelationStruct,
	knownKids map[widget.TreeNodeID][]widget.TreeNodeID,
	knownBits map[widget.TreeNodeID]azure.RelationStruct,
	thenWindow *fyne.Window) *fyne.Container {
	relationshipList := createRelationshipList(
		selectedRelations,
		knownKids,
		knownBits)
	var returningContainer *fyne.Container
	returningContainer = container.NewBorder(
		widget.NewToolbar(
			widget.NewToolbarAction(
				theme.ContentAddIcon(),
				func() {
					addRelWindow := addWindowFor("Add Relationship", 500, 250)
					objectType := widget.NewSelectEntry([]string{
						"Actor",
						"Application Service",
						"Business Service",
						"Capability",
						"Constraint",
						"Data Entity",
						"Interface",
						"Location",
						"Logical Application Component",
						"Logical Data Component",
						"Logical Technology Component",
						"Organization Unit",
						"Physical Application Component",
						"Physical Data Component",
						"Physical Technology Component",
						"Physical Technology Group",
						"Principle",
						"Process",
						"Product",
						"Requirement",
						"Risk",
						"Role",
						"Technology Service",
					})
					objectTypesList := map[string]string{
						"Actor":                          "445f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Application Service":            "bc5f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Business Service":               "73d7af8c-5e52-ea11-a94c-28187852a561",
						"Capability":                     "265f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Constraint":                     "535f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Data Entity":                    "625f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Interface":                      "a3b624e4-b642-ea11-a601-28187852aafd",
						"Location":                       "cb5f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Logical Application Component":  "7cb624e4-b642-ea11-a601-28187852aafd",
						"Logical Data Component":         "96b624e4-b642-ea11-a601-28187852aafd",
						"Logical Technology Component":   "070714ec-b642-ea11-a601-28187852aafd",
						"Organization Unit":              "b0b624e4-b642-ea11-a601-28187852aafd",
						"Physical Application Component": "6fb624e4-b642-ea11-a601-28187852aafd",
						"Physical Data Component":        "37395db8-2eef-e811-9f2b-00155d26bcf8",
						"Physical Technology Component":  "140714ec-b642-ea11-a601-28187852aafd",
						"Physical Technology Group":      "5171e716-436d-ee11-9942-00224895c2e5",
						"Principle":                      "70395db8-2eef-e811-9f2b-00155d26bcf8",
						"Process":                        "7f395db8-2eef-e811-9f2b-00155d26bcf8",
						"Product":                        "8e395db8-2eef-e811-9f2b-00155d26bcf8",
						"Requirement":                    "9d395db8-2eef-e811-9f2b-00155d26bcf8",
						"Risk":                           "b65f6dbe-2eef-e811-9f2b-00155d26bcf8",
						"Role":                           "243a5db8-2eef-e811-9f2b-00155d26bcf8",
						"Technology Service":             "52395db8-2eef-e811-9f2b-00155d26bcf8",
					}
					relationshipSelect := widget.NewSelectEntry([]string{})
					relationshipTypesList := map[string]struct {
						id               string
						typepair         string
						leadobjecttypeid string
					}{}
					objectSelect := widget.NewSelectEntry([]string{})
					objectSelectList := map[string]string{}
					addRelWindow.SetContent(
						container.NewBorder(
							nil,
							widget.NewToolbar(
								widget.NewToolbarAction(
									theme.DocumentSaveIcon(),
									func() {
										leadObject := basics.ObjectId
										memberObject := objectSelectList[objectSelect.Text]
										if basics.ObjectType.Id != relationshipTypesList[relationshipSelect.Text].leadobjecttypeid {
											fmt.Printf("SWAPPING %s %s", basics.ObjectType.Id, relationshipTypesList[relationshipSelect.Text].leadobjecttypeid)
											leadObject = objectSelectList[objectSelect.Text]
											memberObject = basics.ObjectId

										}
										path := "/odata/Relationships"
										query := ``
										body := fmt.Sprintf(
											`{"RelationshipTypeId":"%s",
											"ModelId":"%s",
											"RelationshipTypePairId":"%s",
											"LeadModelItemId":"%s",
											"MemberModelItemId":"%s"}`,
											relationshipTypesList[relationshipSelect.Text].id,
											azure.BaselineArchitectureModel,
											relationshipTypesList[relationshipSelect.Text].typepair,
											leadObject,
											memberObject,
										)
										fmt.Printf("Saving %s\n", body)
										mep, err := az.CallRestEndpoint(
											"POST",
											path,
											[]byte(body),
											query)
										if err != nil {
											dialog.ShowInformation(
												"Failed to save",
												err.Error(),
												*thenWindow,
											)
										} else {
											_, err2 := io.ReadAll(mep)
											if err2 != nil {
												dialog.ShowInformation(
													"Failed to save",
													err2.Error(),
													*thenWindow,
												)
											} else {
												dialog.ShowInformation(
													"Save success",
													"",
													*thenWindow,
												)
											}
										}
										addRelWindow.Close()
									},
								),
								widget.NewToolbarAction(
									theme.CancelIcon(),
									func() {
										addRelWindow.Close()
									},
								),
							),
							nil,
							nil,
							container.New(
								layout.NewFormLayout(),
								widget.NewLabel("Object type"),
								container.NewBorder(
									nil,
									nil,
									nil,
									widget.NewButtonWithIcon(
										"",
										theme.ListIcon(),
										func() {
											mike := az.GetRelationTypesForObjectType(
												basics.ObjectType.Id,
												objectTypesList[objectType.Text],
											)
											selects := []string{}
											for id, obj := range mike {
												relationshipTypesList[obj.Name] = struct {
													id               string
													typepair         string
													leadobjecttypeid string
												}{id, obj.RelationshipTypePairs[0].RelationshipTypePairId, obj.RelationshipTypePairs[0].LeadObjectTypeId}
												selects = append(selects, obj.Name)
											}
											relationshipSelect.SetOptions(selects)
										},
									),
									objectType,
								),
								widget.NewLabel("Relationship"),
								relationshipSelect,

								widget.NewLabel("With object"),
								container.NewBorder(
									nil,
									nil,
									nil,

									widget.NewButtonWithIcon(
										"",
										theme.ListIcon(),
										func() {
											az.FindMeInTypeThen(
												objectSelect.Text,
												objectTypesList[objectType.Text],
												func(finds []azure.FindStruct) {
													returns := []string{}
													objectSelectList = map[string]string{}
													for _, x := range finds {
														returns = append(returns, x.Name)
														objectSelectList[x.Name] = x.ObjectId
													}
													objectSelect.SetOptions(returns)
												})
											mike := az.GetRelationTypesForObjectType(
												basics.ObjectType.Id,
												objectTypesList[objectType.Text],
											)
											selects := []string{}
											for id, obj := range mike {
												relationshipTypesList[obj.Name] = struct {
													id               string
													typepair         string
													leadobjecttypeid string
												}{id, obj.RelationshipTypePairs[0].RelationshipTypePairId, obj.RelationshipTypePairs[0].LeadObjectTypeId}
												selects = append(selects, obj.Name)
											}
											relationshipSelect.SetOptions(selects)
										},
									),
									objectSelect,
								),
							),
						),
					)
					addRelWindow.Show()
				},
			),
			widget.NewToolbarAction(
				theme.ContentRemoveIcon(),
				func() {
					// Prompt for confirmation
					dialog.ShowConfirm(
						"Really delete?",
						"Do you really want to delete the selected entries?",
						func(ok bool) {
							if !ok {
								return
							}
							// If yes, delete
							for _, x := range selectedRelations {
								fmt.Printf("Deleting relationship %s\n", x.RelationshipId)
							}
						},
						*thenWindow,
					)
					fmt.Printf("Remove Relationship(s)")
				},
			),
			widget.NewToolbarAction(
				theme.ColorPaletteIcon(),
				func() {
					filename := widget.NewEntry()
					dialog.ShowForm(
						"Save diagram",
						"Save",
						"Don't",
						[]*widget.FormItem{widget.NewFormItem("Filename", filename)},
						func(save bool) {
							if !save {
								return
							}
							fileName := filepath.Join(getSavePath(), filepath.Base(filename.Text))
							fo, err := os.Create(fileName)
							if err != nil {
								panic(err)
							}
							// close fo on exit and check for its returned error
							defer func() {
								if err := fo.Close(); err != nil {
									panic(err)
								}
							}()
							fo.WriteString(PlantUMLStart)
							alreadyDrawn := map[string]string{}
							alreadyDrawn = map[string]string{
								basics.ObjectId: nameToToken(&alreadyDrawn, basics.Name),
							}
							relationships := map[string]relationshipStruct{}
							objects := map[string]objectStruct{}
							addToObjectStruct(&objects, alreadyDrawn[basics.ObjectId], basics.Name, "physical application component")
							for _, x := range selectedRelations {
								leftAlias := ""
								rightAlias := ""
								var y bool
								if leftAlias, y = alreadyDrawn[x.LeadObjectId]; !y {
									leftAlias = nameToToken(&alreadyDrawn, x.LeadObject.Name)
									alreadyDrawn[x.LeadObjectId] = leftAlias
									addToObjectStruct(&objects, alreadyDrawn[x.LeadObjectId], x.LeadObject.Name, x.LeadObject.Type.Name)
								}
								if rightAlias, y = alreadyDrawn[x.MemberObjectId]; !y {
									rightAlias = nameToToken(&alreadyDrawn, x.MemberObject.Name)
									alreadyDrawn[x.MemberObjectId] = rightAlias
									addToObjectStruct(&objects, alreadyDrawn[x.MemberObjectId], x.MemberObject.Name, x.MemberObject.Type.Name)
								}
								addRelationship(
									&relationships,
									&objects,
									leftAlias,
									x.LeadObject,
									rightAlias,
									x.MemberObject,
									x.RelationshipId,
									x.RelationshipType.LeadToMemberDirection,
								)
							}
							for _, x := range objects {
								fo.WriteString(drawObject(x))
							}
							for _, x := range relationships {
								fo.WriteString(
									fmt.Sprintf(
										"Rel(%s,%s,\"%s\")\n",
										x.leftAlias,
										x.rightAlias,
										x.relationshipName,
									))
							}
							fo.WriteString(PlantUMLEnd)
							dialog.ShowInformation(
								"Saved",
								fmt.Sprintf("Saved the diagram to %s", fileName),
								*thenWindow,
							)
						},
						*thenWindow,
					)
				},
			),
			widget.NewToolbarAction(
				theme.ViewRefreshIcon(),
				func() {
					rels := az.FindRelations(basics.ObjectId)
					knownKids = map[widget.TreeNodeID][]widget.TreeNodeID{}
					knownBits = map[widget.TreeNodeID]azure.RelationStruct{}
					for _, x := range rels {
						knownKids[""] = append(knownKids[""], x.RelationshipId)
						knownBits[x.RelationshipId] = x
					}
					returningContainer.Objects[0] = createRelationshipList(
						selectedRelations,
						knownKids,
						knownBits)
				},
			)),
		nil,
		nil,
		container.NewGridWithColumns(1, widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel("")),
		relationshipList,
	)
	return returningContainer
}

var PlantUMLStart = "@startuml Solution Context\n!include https://raw.githubusercontent.com/colinmo/iserver-diagram/main/togaf/togaf-full.puml\n"
var PlantUMLEnd = "@enduml"

/* Let people press enter to submit a search */
type enterEntry struct {
	widget.Entry
	searchButton *widget.Button
}

func (e *enterEntry) onEnter() {
	position := fyne.Position{X: 0, Y: 0}
	pointEvent := fyne.PointEvent{AbsolutePosition: position, Position: position}
	e.searchButton.Tapped(&pointEvent)
}

func newEnterEntryWithData(text binding.String, button *widget.Button) *enterEntry {
	entry := &enterEntry{
		searchButton: button,
	}
	entry.ExtendBaseWidget(entry)
	entry.Bind(text)
	return entry
}

func (e *enterEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyEnter:
		e.onEnter()
	case fyne.KeyReturn:
		e.onEnter()
	default:
		e.Entry.TypedKey(key)
	}
}

func dateValidator(str string) error {
	if len(str) == 0 {
		return nil
	}
	str = strings.ReplaceAll(str, "/", "-")
	_, not := time.Parse("2006-01-02", str)
	return not
}

/** Model specific layouts **/
func PtcFields() modelFields {
	return modelFields{
		stringValues: map[string]*widget.Entry{
			"Title":                            widget.NewEntry(),
			"Description":                      widget.NewMultiLineEntry(),
			"GU::Information System Custodian": widget.NewEntry(),
			"Supplier":                         widget.NewEntry(),
			"Department":                       widget.NewEntry(),
		},
		selectValues: map[string]*widget.Select{
			"Owner":      widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Domain": widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Information Security Classification": widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Solution Classification":             widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Object Visibility":                   widget.NewSelect([]string{}, func(bob string) {}),
			"Lifecycle Status":                        widget.NewSelect([]string{}, func(bob string) {}),
			"Internal Recommendation":                 widget.NewSelect([]string{}, func(bob string) {}),
			"Operational Importance":                  widget.NewSelect([]string{}, func(bob string) {}),
		},
		radioValues: map[string]*widget.RadioGroup{},
		checkValues: map[string]*widget.CheckGroup{},
		dateValues: map[string]*widget.Entry{
			"Internal: In Development From": widget.NewEntry(),
			"Internal: Live date":           widget.NewEntry(),
			"Internal: Phase Out From":      widget.NewEntry(),
			"Internal: Retirement date":     widget.NewEntry()},
		sections: map[int]sectionStruct{
			0: {
				title: "Basic",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"Name", "string", "Title"}},
					1: {0: {"Description", "string", "Description"}},
					2: {0: {"Domain", "select", "GU::Domain"}},
				},
			},
			1: {
				title: "Roles",
				fields: map[int]map[int]fieldsStruct{
					1: {0: {"Owner (Product Manager)", "select", "Owner"}},
					2: {0: {"Custodian", "string", "GU::Information System Custodian"}},
					3: {0: {"Supplier", "string", "Supplier"}},
					4: {0: {"Department (Business Owner)", "string", "Department"}},
				},
			},
			2: {
				title: "Meta",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"Information Security classification", "select", "GU::Information Security Classification"}},
					1: {0: {"Solution classification", "select", "GU::Solution Classification"}},
					2: {0: {"Visible in applist", "select", "GU::Object Visibility"}},
					4: {0: {"Internal recommendation", "select", "Internal Recommendation"}},
					5: {0: {"Operational importance", "select", "Operational Importance"}},
				},
			},
			3: {
				title: "Dates",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"In development", "date", "Internal: In Development From"}},
					1: {0: {"Live", "date", "Internal: Live date"}},
					2: {0: {"Phasing out", "date", "Internal: Phase Out From"}},
					3: {0: {"Retirement", "date", "Internal: Retirement date"}},
					4: {0: {"Lifecycle Status", "select", "Lifecycle Status"}},
				},
			},
		},
	}
}

func PacFields() modelFields {
	return modelFields{
		stringValues: map[string]*widget.Entry{
			"Title":                            widget.NewEntry(),
			"Description":                      widget.NewMultiLineEntry(),
			"Alias":                            widget.NewEntry(),
			"Links":                            widget.NewMultiLineEntry(),
			"GU::Information System Custodian": widget.NewEntry(),
			"Vendor":                           widget.NewEntry(),
			"Supplier":                         widget.NewEntry(),
			"Department":                       widget.NewEntry(),
			"Approved Usage":                   widget.NewMultiLineEntry(),
			"Serviceability characteristics":   widget.NewEntry(),
			"Conditions & Restrictions":        widget.NewMultiLineEntry(),
		},
		selectValues: map[string]*widget.Select{
			"Application Type":          widget.NewSelect([]string{}, func(bob string) {}),
			"Operational Importance":    widget.NewSelect([]string{}, func(bob string) {}),
			"Deployment Method":         widget.NewSelect([]string{}, func(bob string) {}),
			"Build":                     widget.NewSelect([]string{}, func(bob string) {}),
			"Owner":                     widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Domain":                widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Managed outside of DS": widget.NewSelect([]string{"True", "False"}, func(bob string) {}),
			"GU::Information Security Classification": widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Solution Classification":             widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Object Visibility":                   widget.NewSelect([]string{}, func(bob string) {}),
			"Internal Recommendation":                 widget.NewSelect([]string{}, func(bob string) {}),
			"Standards Class":                         widget.NewSelect([]string{}, func(bob string) {}),
			"Lifecycle Status":                        widget.NewSelect([]string{}, func(bob string) {}),
		},
		radioValues: map[string]*widget.RadioGroup{},
		checkValues: map[string]*widget.CheckGroup{
			"Categories": widget.NewCheckGroup([]string{}, func(bob []string) {}),
		},
		dateValues: map[string]*widget.Entry{
			"Internal: In Development From": widget.NewEntry(),
			"Internal: Live date":           widget.NewEntry(),
			"Internal: Phase Out From":      widget.NewEntry(),
			"Internal: Retirement date":     widget.NewEntry(),
			"Date of Last Release":          widget.NewEntry(),
			"Date of Next Release":          widget.NewEntry(),
			"Vendor: Contained From":        widget.NewEntry(),
			"Vendor: Out of Support":        widget.NewEntry(),
			"Standard Creation Date":        widget.NewEntry(),
			"Last Standard Review Date":     widget.NewEntry(),
			"Next Standard Review Date":     widget.NewEntry(),
			"Standard Retire Date":          widget.NewEntry(),
		},
		sections: map[int]sectionStruct{
			0: {
				title: "Key attributes",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"Name", "string", "Title"}},
					1: {0: {"Description", "string", "Description"}},
					2: {0: {"Domain", "select", "GU::Domain"}},
					3: {0: {"Alias", "string", "Alias"}},
					4: {0: {"Links", "string", "Links"}},
					5: {0: {"Categories", "check", "Categories"}},
					6: {0: {"Department (Requestor)", "string", "Department"},
						1: {"Owner (DS Area)", "select", "Owner"}},
					7: {0: {"Solution classification", "select", "GU::Solution Classification"},
						1: {"Information security classification", "select", "GU::Information Security Classification"}},
					8: {0: {"Vendor", "string", "Vendor"},
						1: {"Supplier", "string", "Supplier"}},
				},
			},
			1: {
				title: "Lifecycle & Roadmap",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"Lifecycle Status", "select", "Lifecycle Status"}},
					1: {0: {"Internal recommendation", "select", "Internal Recommendation"}},
					2: {0: {"Date of Last Release", "date", "Date of Last Release"},
						1: {"Date of Next Release", "date", "Date of Next Release"}},
					3: {0: {"In development", "date", "Internal: In Development From"}, 1: {"Live", "date", "Internal: Live date"}, 2: {"Phasing out", "date", "Internal: Phase Out From"}, 3: {"Retirement", "date", "Internal: Retirement date"}},
					4: {0: {"Vendor Contained From", "date", "Vendor: Contained From"}, 1: {"Vendor Out of Support", "date", "Vendor: Out of Support"}},
					5: {0: {"Serviceability characteristics", "string", "Serviceability characteristics"}},
				},
			},
			2: {
				title: "Standards & Usage",
				fields: map[int]map[int]fieldsStruct{
					0: {0: {"Standards Class", "select", "Standards Class"}},
					1: {0: {"Standard Creation Date", "date", "Standard Creation Date"},
						1: {"Last Standard Review Date", "date", "Last Standard Review Date"},
						2: {"Next Standard Review Date", "date", "Next Standard Review Date"},
						3: {"Standard Retire Date", "date", "Standard Retire Date"}},
					2: {0: {"Approved Usage", "string", "Approved Usage"}},
					3: {0: {"Conditions & Restrictions", "string", "Conditions & Restrictions"}},
					4: {0: {"Application Type", "select", "Application Type"}},
					5: {0: {"Operational Importance", "select", "Operational Importance"}},
					6: {0: {"Deployment Method", "select", "Deployment Method"}},
					7: {0: {"Build", "select", "Build"}},
				},
			},
		},
	}
}

func LacFields() modelFields {
	return modelFields{}
}

/** Generic utility functions **/
func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func getSavePath() string {
	savePath := myApp.Preferences().StringWithFallback("SavePath", "")
	if savePath == "" {
		var err error
		savePath, err = os.UserHomeDir()
		if err != nil {
			savePath = os.TempDir()
		}
		myApp.Preferences().SetString("SavePath", savePath)
	}
	return savePath
}

func PrettyJSONString(str string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return str
	}
	return prettyJSON.String()
}
