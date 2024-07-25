package main

import (
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
	mainWindow := myApp.NewWindow("Loading")
	mainWindow.Resize(fyne.NewSize(400, 600))
	bottom := container.New(
		layout.NewHBoxLayout(),
		widget.NewLabelWithData(status),
		layout.NewSpacer(),
		widget.NewLabelWithData(messages),
	)
	searchEntry := binding.NewString()
	centreContent = container.NewMax()
	// Settings
	pms := widget.NewMultiLineEntry()
	pms.SetText(myApp.Preferences().StringWithFallback("ProductManagers", "[]"))
	savepath := widget.NewEntry()
	savepath.SetText(myApp.Preferences().StringWithFallback("SavePath", ""))
	searchButton := widget.NewButton(
		"Go",
		func() {
			if x, _ := status.Get(); x == "Live" {
				UpdateMessage("Searching...")
				text, _ := searchEntry.Get()
				az.FindMeThen(text, ListAndSelectAThing, &mainWindow)
				UpdateMessage("Ready")
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
							theme.FileIcon(),
							func() {
								fmt.Print("Prompt for Object Type")
								fmt.Print("Open edit window")
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
						myApp.Preferences().SetString("ProductManagers", pms.Text)
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
				widget.NewButtonWithIcon("Load", &fyne.StaticResource{}, func() {}),
				widget.NewLabel("template"),
			)
		},
		func(id int, item fyne.CanvasObject) {
			switch things[id].Type.Name {
			case "Physical Technology Component":
				item.(*fyne.Container).Objects[0].(*widget.Button).SetIcon(resourcePtcPng)
				item.(*fyne.Container).Objects[0].(*widget.Button).SetText("PTC")
			case "Physical Application Component":
				item.(*fyne.Container).Objects[0].(*widget.Button).SetIcon(resourcePacPng)
				item.(*fyne.Container).Objects[0].(*widget.Button).SetText("PAC")
			case "Logical Application Component":
				item.(*fyne.Container).Objects[0].(*widget.Button).SetIcon(resourceLacPng)
				item.(*fyne.Container).Objects[0].(*widget.Button).SetText("LAC")
			default:
				fmt.Printf("Dunno: %s\n", things[id].Type.Name)
			}
			item.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {
				windowTitle := fmt.Sprintf("Details for %s", things[id].Name)
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
				az.FindRelationsThen(things[id].ObjectId, ListRelationsToSelect, &lookupWindow)
				UpdateMessage("Ready")
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
	fields map[int]fieldsStruct
}

type modelFields struct {
	stringValues map[string]*widget.Entry
	selectValues map[string]*widget.Select
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
	json.Unmarshal([]byte(myApp.Preferences().StringWithFallback("ProductManagers", "[]")), &allFields.selectValues["Owner"].Options)
	allFields.stringValues["Title"].SetText(basics.Name)
	selectedRelations := map[string]azure.RelationStruct{}
	for i := range allFields.dateValues {
		allFields.dateValues[i].Validator = dateValidator
	}
	isString := func(str string) bool { _, x := allFields.stringValues[str]; return x }
	isSelect := func(str string) bool { _, x := allFields.selectValues[str]; return x }
	isDate := func(str string) bool { _, x := allFields.dateValues[str]; return x }
	for _, x := range basics.AttributeValues {
		switch {
		case isString(x.AttributeName):
			allFields.stringValues[x.AttributeName].SetText(x.StringValue)
		case isSelect(x.AttributeName):
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
		case isDate(x.AttributeName):
			allFields.dateValues[x.AttributeName].SetText(strings.Replace(x.StringValue, "T00:00:00Z", "", 1))
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

func makeEditPage(allFields modelFields, relationshipWindow *fyne.Container, thisWindow *fyne.Window) *fyne.Container {
	box := container.NewVBox()
	for i := 0; i < len(allFields.sections); i++ {
		baseform := widget.NewForm()
		sec := allFields.sections[i]
		for j := 0; j < len(sec.fields); j++ {
			fld := sec.fields[j]
			switch fld.fieldindex {
			case "string":
				baseform.AppendItem(widget.NewFormItem(fld.label, allFields.stringValues[fld.valuesIndex]))
			case "select":
				baseform.AppendItem(widget.NewFormItem(fld.label, allFields.selectValues[fld.valuesIndex]))
			case "date":
				baseform.AppendItem(
					widget.NewFormItem(
						fld.label,
						container.NewBorder(
							nil, nil, nil,
							widget.NewButtonWithIcon(
								"",
								mywidge.CalendarResource,
								func() {
									var deepdeep dialog.Dialog
									deepdeep = dialog.NewCustom(
										"Change date",
										"Nevermind",
										mywidge.CreateDatePicker(
											stringToDate(allFields.dateValues[fld.valuesIndex].Text),
											&deepdeep,
											allFields.dateValues[fld.valuesIndex]),
										*thisWindow,
									)
									deepdeep.Show()
								},
							),
							allFields.dateValues[fld.valuesIndex])))
			}
		}
		subsection := container.NewBorder(
			widget.NewLabelWithStyle(sec.title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			baseform)
		box.Objects = append(box.Objects, subsection)
	}
	box.Objects = append(box.Objects, relationshipWindow)
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
					az.FindRelationsThen(meps[1], ListRelationsToSelect, &lookupWindow)
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
	switch object.otype {
	case "Physical Application Component":
		return fmt.Sprintf(
			"System(%s,\"%s\",\"\",\"\",$type=\"PAC\")\n",
			object.alias,
			object.name,
		)
	case "Physical Technology Component":
		return fmt.Sprintf(
			"System(%s,\"%s\",\"\",\"\",$type=\"PTC\")\n",
			object.alias,
			object.name,
		)
	case "Location":
		children := strings.Builder{}
		for _, x := range object.children {
			children.WriteString(drawObject(x))
		}
		return fmt.Sprintf(
			"Enterprise_Boundary(%s,\"%s\",\"\") {\n%s}\n",
			object.alias,
			object.name,
			children.String(),
		)
	case "Logical Application Component":
		children := strings.Builder{}
		for _, x := range object.children {
			children.WriteString(drawObject(x))
		}
		return fmt.Sprintf(
			"System_Boundary(%s,\"%s\",\"\") {\n%s}\n",
			object.alias,
			object.name,
			children.String(),
		)
	case "Capability":
		return fmt.Sprintf(
			"System(%s,\"%s\",\"\",\"\",$type=\"CAP\")\n",
			object.alias,
			object.name,
		)
	default:
		return fmt.Sprintf(
			"System(%s,\"%s\",\"\",\"\",$type=\"%v\")\n",
			object.alias,
			object.name,
			object.otype,
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

func createRelationshipWindow(
	basics azure.IServerObjectStruct,
	selectedRelations map[string]azure.RelationStruct,
	knownKids map[widget.TreeNodeID][]widget.TreeNodeID,
	knownBits map[widget.TreeNodeID]azure.RelationStruct,
	thenWindow *fyne.Window) *fyne.Container {
	relationshipList := widget.NewTree(
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
	return container.NewBorder(
		widget.NewLabelWithStyle("Relationships", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewToolbar(
			widget.NewToolbarAction(
				theme.ContentAddIcon(),
				func() {
					addRelWindow := addWindowFor("Add Relationship", 500, 250)
					objectType := widget.NewSelectEntry([]string{
						"Capability",
						"Actor",
						"Constraint",
						"Data Entity",
						"Application Service",
						"Location",
						"Physical Data Component",
						"Technology Service",
						"Principle",
						"Process",
						"Product",
						"Requirement",
						"Role",
						"Risk",
						"Physical Technology Group",
						"Business Service",
						"Physical Application Component",
						"Logical Application Component",
						"Logical Data Component",
						"Interface",
						"Organization Unit",
						"Logical Technology Component",
						"Physical Technology Component",
					})
					objectTypesList := map[string]string{
						"Capability":                     "265f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Actor":                          "445f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Constraint":                     "535f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Data Entity":                    "625f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Application Service":            "bc5f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Location":                       "cb5f5bb2-2eef-e811-9f2b-00155d26bcf8",
						"Physical Data Component":        "37395db8-2eef-e811-9f2b-00155d26bcf8",
						"Technology Service":             "52395db8-2eef-e811-9f2b-00155d26bcf8",
						"Principle":                      "70395db8-2eef-e811-9f2b-00155d26bcf8",
						"Process":                        "7f395db8-2eef-e811-9f2b-00155d26bcf8",
						"Product":                        "8e395db8-2eef-e811-9f2b-00155d26bcf8",
						"Requirement":                    "9d395db8-2eef-e811-9f2b-00155d26bcf8",
						"Role":                           "243a5db8-2eef-e811-9f2b-00155d26bcf8",
						"Risk":                           "b65f6dbe-2eef-e811-9f2b-00155d26bcf8",
						"Physical Technology Group":      "5171e716-436d-ee11-9942-00224895c2e5",
						"Business Service":               "73d7af8c-5e52-ea11-a94c-28187852a561",
						"Physical Application Component": "6fb624e4-b642-ea11-a601-28187852aafd",
						"Logical Application Component":  "7cb624e4-b642-ea11-a601-28187852aafd",
						"Logical Data Component":         "96b624e4-b642-ea11-a601-28187852aafd",
						"Interface":                      "a3b624e4-b642-ea11-a601-28187852aafd",
						"Organization Unit":              "b0b624e4-b642-ea11-a601-28187852aafd",
						"Logical Technology Component":   "070714ec-b642-ea11-a601-28187852aafd",
						"Physical Technology Component":  "140714ec-b642-ea11-a601-28187852aafd",
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
										mep, _ := az.CallRestEndpoint(
											"POST",
											path,
											[]byte(body),
											query)
										bytemep, err := io.ReadAll(mep)
										if err != nil {
											log.Fatalf("failed to read io.Reader %v\n", err)
										}
										fmt.Printf("%s\n\n", body)
										fmt.Printf("%s\n\n%v", string(bytemep), err)
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
										theme.MailComposeIcon(),
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
										theme.MailComposeIcon(),
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
												}{id, obj.RelationshipTypePairs[0].RelationshipTypePairId, obj.RelationshipTypePairs[0].RelationshipTypePairId}
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
					// Pop up window
					// Get the valid type relations
					// Fill the prompt for relationship type
					// Search for an object of the type
					// Save relationship
					fmt.Printf("Add Relationship")
				},
			),
			widget.NewToolbarAction(
				theme.ContentRemoveIcon(),
				func() {
					// Prompt for confirmation
					// If yes, delete
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
							savePath := myApp.Preferences().StringWithFallback("SavePath", "")
							if savePath == "" {
								var err error
								savePath, err = os.UserHomeDir()
								if err != nil {
									savePath = os.TempDir()
								}
								myApp.Preferences().SetString("SavePath", savePath)
							}
							fileName := filepath.Join(savePath, filepath.Base(filename.Text))
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
							addToObjectStruct(&objects, alreadyDrawn[basics.ObjectId], basics.Name, "PAC")
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
			)),
		nil,
		container.NewGridWithColumns(1, widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel(""), widget.NewLabel("")),
		relationshipList)
}

var PlantUMLStart = "@startuml Solution Context\n!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml\n!define DEVICONS https://raw.githubusercontent.com/tupadr3/plantuml-icon-font-sprites/master/devicons\n!define FONTAWESOME https://raw.githubusercontent.com/tupadr3/plantuml-icon-font-sprites/master/font-awesome-5\n!include DEVICONS/angular.puml\n!include DEVICONS/java.puml\n!include DEVICONS/msql_server.puml\n!include FONTAWESOME/users.puml\nSetDefaultLegendEntries(\"\")\nLAYOUT_WITH_LEGEND()\n"
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

func stringToDate(str string) time.Time {
	if len(str) == 0 {
		return time.Now()
	}
	str = strings.ReplaceAll(str, "/", "-")
	dt, e := time.Parse("2006-01-02", str)
	if e == nil {
		return dt
	}
	return time.Now()
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
			"GU::Review Bodies":                widget.NewEntry(),
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
		dateValues: map[string]*widget.Entry{
			"Internal: In Development From": widget.NewEntry(),
			"Internal: Live date":           widget.NewEntry(),
			"Internal: Phase Out From":      widget.NewEntry(),
			"Internal: Retirement date":     widget.NewEntry()},
		sections: map[int]sectionStruct{
			0: {
				title: "Basic",
				fields: map[int]fieldsStruct{
					0: {"Name", "string", "Title"},
					1: {"Description", "string", "Description"},
					2: {"Domain", "select", "GU::Domain"},
				},
			},
			1: {
				title: "Roles",
				fields: map[int]fieldsStruct{
					1: {"Owner (Product Manager)", "select", "Owner"},
					2: {"Custodian", "string", "GU::Information System Custodian"},
					3: {"Supplier", "string", "Supplier"},
					4: {"Department (Business Owner)", "string", "Department"},
				},
			},
			2: {
				title: "Meta",
				fields: map[int]fieldsStruct{
					0: {"Information Security classification", "select", "GU::Information Security Classification"},
					1: {"Solution classification", "select", "GU::Solution Classification"},
					2: {"Visible in applist", "select", "GU::Object Visibility"},
					3: {"Review", "string", "GU::Review Bodies"},
					4: {"Internal recommendation", "select", "Internal Recommendation"},
					5: {"Operational importance", "select", "Operational Importance"},
				},
			},
			3: {
				title: "Dates",
				fields: map[int]fieldsStruct{
					0: {"In development", "date", "Internal: In Development From"},
					1: {"Live", "date", "Internal: Live date"},
					2: {"Phasing out", "date", "Internal: Phase Out From"},
					3: {"Retirement", "date", "Internal: Retirement date"},
					4: {"Lifecycle Status", "select", "Lifecycle Status"},
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
			"Owner (Legacy)":                   widget.NewEntry(),
			"GU::Information System Custodian": widget.NewEntry(),
			"GU::Review Bodies":                widget.NewEntry(),
			"Supplier":                         widget.NewEntry(),
			"Department":                       widget.NewEntry(),
		},
		selectValues: map[string]*widget.Select{
			"Owner":      widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Domain": widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Managed outside of DS": widget.NewSelect(
				[]string{"True", "False"},
				func(bob string) {},
			),
			"GU::Information Security Classification": widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Solution Classification":             widget.NewSelect([]string{}, func(bob string) {}),
			"GU::Object Visibility":                   widget.NewSelect([]string{}, func(bob string) {}),
			"Lifecycle Status":                        widget.NewSelect([]string{}, func(bob string) {}),
			"Internal Recommendation":                 widget.NewSelect([]string{}, func(bob string) {}),
			"Operational Importance":                  widget.NewSelect([]string{}, func(bob string) {}),
		},
		dateValues: map[string]*widget.Entry{
			"Internal: In Development From": widget.NewEntry(),
			"Internal: Live date":           widget.NewEntry(),
			"Internal: Phase Out From":      widget.NewEntry(),
			"Internal: Retirement date":     widget.NewEntry(),
		},
		sections: map[int]sectionStruct{
			0: {
				title: "Basic",
				fields: map[int]fieldsStruct{
					0: {"Name", "string", "Title"},
					1: {"Description", "string", "Description"},
					2: {"Domain", "select", "GU::Domain"},
				},
			},
			1: {
				title: "Roles",
				fields: map[int]fieldsStruct{
					0: {"Owner (Product Manager)", "select", "Owner"},
					1: {"Owner (Legacy)", "string", "Owner (Legacy)"},
					2: {"Custodian", "string", "GU::Information System Custodian"},
					3: {"Supplier", "string", "Supplier"},
					4: {"Department (Business Owner)", "string", "Department"},
				},
			},
			2: {
				title: "Meta",
				fields: map[int]fieldsStruct{
					0: {"Information security classification", "select", "GU::Information Security Classification"},
					1: {"Solution classification", "select", "GU::Solution Classification"},
					2: {"Visible in applist", "select", "GU::Object Visibility"},
					3: {"Review", "string", "GU::Review Bodies"},
					4: {"Internal recommendation", "select", "Internal Recommendation"},
					5: {"Operational importance", "select", "Operational Importance"},
				},
			},
			3: {
				title: "Dates",
				fields: map[int]fieldsStruct{
					0: {"In development", "date", "Internal: In Development From"},
					1: {"Live", "date", "Internal: Live date"},
					2: {"Phasing out", "date", "Internal: Phase Out From"},
					3: {"Retirement", "date", "Internal: Retirement date"},
					4: {"Lifecycle Status", "select", "Lifecycle Status"},
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
