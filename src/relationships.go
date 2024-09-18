package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	azure "vonexplaino.com/m/v2/vondiagram/azure"
)

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

					handleObjectTypeChange := func(ch string) {
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
					}
					objectType.OnChanged = handleObjectTypeChange
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
								objectType,
								widget.NewLabel("Relationship"),
								relationshipSelect,
								widget.NewLabel("With object"),
								container.NewBorder(
									nil,
									nil,
									nil,
									widget.NewButtonWithIcon(
										"",
										theme.SearchIcon(),
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

							errors := []string{}
							for _, x := range selectedRelations {
								err := az.DeleteARelationship(x.RelationshipId)
								if err != nil {
									errors = append(errors, err.Error())
								}
							}
							if len(errors) == 0 {
								dialog.ShowInformation("Successful deletion", "", *thenWindow)
							} else {
								dialog.ShowError(fmt.Errorf("the following messages were returned from the endpoint:\n"+strings.Join(errors, "\n")), *thenWindow)
							}
						},
						*thenWindow,
					)
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
							if len(fileName) < 5 || fileName[len(fileName)-5:] != "puml" {
								fileName = fileName + ".puml"
							}
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
