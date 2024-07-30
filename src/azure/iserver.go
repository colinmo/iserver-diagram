package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	fyne "fyne.io/fyne/v2"
	"github.com/xuri/excelize/v2"
)

type FindStruct struct {
	Name     string `json:"Name"`
	ObjectId string `json:"ObjectId"`
	Type     struct {
		Name string `json:"Name"`
	} `json:"ObjectType"`
}

var ValidChoices = map[string]map[string]string{}

var BaselineArchitectureModel = "0bb71446-f140-ea11-a601-28187852aafd"

type ValuesValue struct {
	AttributeConfigurationChoiceId string `json:"AttributeConfigurationChoiceId,omitempty"`
	Value                          string `json:"Value,omitempty"`
	Url                            string `json:"Url,omitempty"`
	DisplayValue                   string `json:"DisplayValue,omitempty"`
}

type AttributeValue struct {
	StringValue   string        `json:"StringValue,omitempty"`
	AttributeId   string        `json:"AttributeId,omitempty"`
	AttributeName string        `json:"AttributeName,omitempty"`
	DataType      string        `json:"@odata.type,omitempty"`
	Values        []ValuesValue `json:"Values,omitempty"`
}
type SaveObject struct {
	Name            string      `json:"Name"`
	ObjectTypeId    string      `json:"ObjectTypeId"`
	ModelId         string      `json:"ModelId"`
	AttributeValues []SaveValue `json:"AttributeValues"`
}
type SaveValue struct {
	AttributeName     string        `json:"AttributeName"`
	AttributeCategory string        `json:"AttributeCategory"`
	TextValue         string        `json:"TextValue,omitempty"`
	DecimalValue      float64       `json:"DecimalValue,omitempty"`
	DateTimeValue     string        `json:"DateTimeValue,omitempty"`
	BooleanValue      bool          `json:"BooleanValue,omitempty"`
	ChoiceValues      []ValuesValue `json:"ChoiceValues,omitempty"`
	Values            []ValuesValue `json:"Values,omitempty"`
}
type IServerObjectStruct struct {
	Name            string           `json:"Name"`
	ObjectId        string           `json:"ObjectId"`
	AttributeValues []AttributeValue `json:"AttributeValues"`
	ObjectType      struct {
		Name string `json:"Name"`
		Id   string `json:"ObjectTypeId"`
	} `json:"ObjectType"`
}

type RelationStruct struct {
	RelationshipId   string `json:"RelationshipId"`
	LeadObjectId     string `json:"LeadObjectId"`
	MemberObjectId   string `json:"MemberObjectId"`
	RelationshipType struct {
		Name                  string `json:"Name"`
		RelationshipTypeId    string `json:"RelationshipTypeId"`
		LeadToMemberDirection string `json:"LeadToMemberDirection"`
	} `json:"RelationshipType"`
	LeadObject   FindStruct `json:"LeadObject"`
	MemberObject FindStruct `json:"MemberObject"`
}

type laterLongUpdate func([]FindStruct, *fyne.Window)
type laterRelationUpdate func(IServerObjectStruct, []RelationStruct, *fyne.Window)
type laterStringList func(map[string][]string, *fyne.Window)
type laterDomainOwned func(map[string][]IServerObjectStruct, fyne.Window)

var ImportantFields = map[string][]string{
	"GEN": {
		"Name",
		"Description",
	},
	"PAC": {
		"Alias",
		"Description",
		"Links",
		"Categories",
		"Owner",
		"GU::Domain",
		"Department",
		"GU::Solution Classification",
		"GU::Information Security Classification",
		"Vendor",
		"Supplier",
		"Application Type",
		"Operational Importance",
		"Deployment Method",
		"Build",

		"Lifecycle Status",
		"Internal Recommendation",
		"Date of Last Release",
		"Date of Next Release",
		"Internal: In Development From",
		"Internal: Live Date",
		"Internal: Phase Out From",
		"Internal: Retirement Date",
		"Vendor: Contained From",
		"Vendor: Out of Support",

		"Standard Class",
		"Standard Creation Date",
		"Last Standard Review Date",
		"Next Standard Review Date",
		"Standard Retire Date",
		"Approved Usage",
		"Conditions & Restrictions",

		"GU::Managed Outside Of DS",
		"GU::Object Visibility",
		"Serviceability characteristics",
	},
}

// Simple find over iServer components, looking for the specified string
// Focuses on PAC, PTC, and LAC
func (a *AzureAuth) FindMeThen(lookFor string, putInto laterLongUpdate, thenWindow *fyne.Window) {
	toReturn := []FindStruct{}

	type objects struct {
		Value    []FindStruct `json:"value"`
		NextLink string       `json:"@odata.nextLink"`
	}

	path := "/odata/Objects"
	query := strings.ReplaceAll(
		fmt.Sprintf(
			`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('ObjectId','Name','ObjectType'))&$filter=Model/Name eq 'Baseline Architecture' and ObjectType/Name in ('Physical Application Component','Physical Technology Component','Logical Application Component') and indexOf(tolower(Name),'%s') gt -1`,
			strings.ToLower(lookFor)),
		" ",
		"%20")
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		toReturn = append(toReturn, oneCall.Value...)
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}

	putInto(toReturn, thenWindow)
}

func (a *AzureAuth) GetImportantFields(id, typeofobject string) IServerObjectStruct {
	toReturn := IServerObjectStruct{}

	query := `$expand=` + url.QueryEscape(`ObjectType($select=Name,ObjectTypeId),AttributeValues($select=StringValue,AttributeName,AttributeId;$filter=AttributeName in ("`+strings.Join(ImportantFields[typeofobject], `","`)+`"))`)

	path := fmt.Sprintf("/odata/Objects(%s)", id)
	mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
	if err != nil {
		log.Fatalf("failed to call endpoint %v\n", err)
	}
	defer mep.Close()
	bytemep, err := io.ReadAll(mep)
	json.Unmarshal(bytemep, &toReturn)
	if err != nil {
		log.Fatalf("failed to read io.Reader %v\n", err)
	}

	return toReturn
}

func (a *AzureAuth) SaveObjectFields(
	id string,
	objectName string,
	stringValues map[string]string,
	selectValues map[string]string,
	dateValues map[string]string,
) (bool, string, string) {
	saveValues := SaveObject{}
	saveValues.Name = stringValues["Title"]
	saveValues.ModelId = BaselineArchitectureModel
	switch objectName {
	case "Physical Application Component":
		saveValues.ObjectTypeId = "6fb624e4-b642-ea11-a601-28187852aafd"
	case "Physical Technology Component":
		saveValues.ObjectTypeId = "140714ec-b642-ea11-a601-28187852aafd"
	case "Logical Application Component":
		saveValues.ObjectTypeId = "7cb624e4-b642-ea11-a601-28187852aafd"
	}
	fmt.Printf("Date values %v\n", dateValues)
	// Name is special
	saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
		AttributeName:     "Name",
		AttributeCategory: "Text",
		TextValue:         stringValues["Title"],
	})
	delete(stringValues, "Title")

	// So is links
	LinkValues := []ValuesValue{}
	re := regexp.MustCompile(`^\s*(.*?)\s*\((.*)\)$\s*`)
	stringValues["Links"] = strings.ReplaceAll(stringValues["Links"], "\n", ",")
	for _, e := range strings.Split(stringValues["Links"], ",") {
		bits := re.FindStringSubmatch(e)
		if len(bits) > 2 {
			LinkValues = append(LinkValues, ValuesValue{
				Url:          bits[2],
				DisplayValue: bits[1],
			})
		}
	}
	saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
		AttributeName:     "Links",
		AttributeCategory: "Hyperlink",
		Values:            LinkValues,
	})
	delete(stringValues, "Links")

	// As is Categories
	CategoryValues := []ValuesValue{}
	for _, e := range strings.Split(selectValues["Categories"], ",") {
		CategoryValues = append(CategoryValues, ValuesValue{
			Value:                          e,
			AttributeConfigurationChoiceId: ValidChoices["Categories"][e],
		})
	}
	saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
		AttributeName:     "Categories",
		AttributeCategory: "Choice",
		ChoiceValues:      CategoryValues,
	})
	delete(selectValues, "Categories")

	// Generics
	for i, e := range stringValues {
		saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
			AttributeName:     i,
			AttributeCategory: "Text",
			TextValue:         e,
		})
	}
	saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
		AttributeName:     "Owner",
		AttributeCategory: "Text",
		TextValue:         selectValues["Owner"],
	})
	delete(selectValues, "Owner")
	_, y := selectValues["GU::Managed outside of DS"]
	if y {
		saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
			AttributeName:     "GU::Managed outside of DS",
			AttributeCategory: "TrueFalse",
			BooleanValue:      selectValues["GU::Managed outside of DS"] == "True",
		})
		delete(selectValues, "GU::Managed outside of DS")
	}
	for i, e := range selectValues {
		saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
			AttributeName:     i,
			AttributeCategory: "Choice",
			ChoiceValues:      []ValuesValue{{Value: e, AttributeConfigurationChoiceId: ValidChoices[i][e]}},
		})

	}
	for i, e := range dateValues {
		saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
			AttributeName:     i,
			AttributeCategory: "DateTime",
			DateTimeValue:     e,
		})

	}
	x, err := json.Marshal(saveValues)
	fmt.Printf("Saving as %s\n", string(x))
	if err == nil {
		var mep io.ReadCloser
		if id == "" {
			// Create
			path := "/odata/Objects"
			query := ``
			mep, _ = a.CallRestEndpoint("POST", path, x, query)
		} else {
			// Update
			path := fmt.Sprintf("/odata/Objects(%s)", id)
			query := ``
			mep, _ = a.CallRestEndpoint("PATCH", path, x, query)
		}
		defer mep.Close()
		toReturn := struct {
			Success  bool `json:"success"`
			Messages []struct {
				Message string `json:"message"`
			} `json:"messages"`
			SuccessMessage struct {
				MessageDefinition struct {
					ObjectId string `json:"ObjectId"`
				} `json:"MessageDefinition"`
			} `json:"SuccessMessage"`
		}{}
		bytemep, err := io.ReadAll(mep)
		fmt.Printf("%s", bytemep)
		json.Unmarshal(bytemep, &toReturn)
		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		if len(toReturn.Messages) == 0 {
			toReturn.Messages = append(toReturn.Messages, struct {
				Message string `json:"message"`
			}{Message: ""})
		}
		returnMessages := []string{}
		for _, x := range toReturn.Messages {
			returnMessages = append(returnMessages, x.Message)
		}
		return toReturn.Success, strings.Join(returnMessages, "\n"), toReturn.SuccessMessage.MessageDefinition.ObjectId
	}
	return false, "Big ol' json packing failure", ""
}

func (a *AzureAuth) FindRelations(id string) []RelationStruct {
	toReturn := []RelationStruct{}

	type objects struct {
		Value    []RelationStruct `json:"value"`
		NextLink string           `json:"@odata.nextLink"`
	}

	path := "/odata/Relationships"
	query := fmt.Sprintf(
		`includeIntersectional=false&%%24select=RelationshipId%%2CLeadObjectId%%2CMemberObjectId%%2CLeadObject%%2CMemberObject&%%24expand=RelationshipType(%%24select%%3DName%%2CLeadToMemberDirection)%%2CLeadObject(%%24select%%3DName%%2CObjectId%%2CObjectType%%3B%%24expand%%3DObjectType(%%24select%%3DName))%%2CMemberObject(%%24select%%3DName%%2CObjectId%%2CObjectType%%3B%%24expand%%3DObjectType(%%24select%%3DName))&%%24filter=LeadObjectId%%20eq%%20%s%%20or%%20MemberObjectId%%20eq%%20%s`,
		id,
		id,
	)
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		toReturn = append(toReturn, oneCall.Value...)
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	return toReturn
}

func (a *AzureAuth) FindRelationsThen(id string, typeofobject string, putInto laterRelationUpdate, thenWindow *fyne.Window) {
	putInto(a.GetImportantFields(id, typeofobject), a.FindRelations(id), thenWindow)
}

func (a *AzureAuth) GetProductManagersThen(department string, putInto laterStringList, thenWindow *fyne.Window) {
	toReturn := map[string][]string{}

	type objects struct {
		Value    []IServerObjectStruct `json:"value"`
		NextLink string                `json:"@odata.nextLink"`
	}

	path := "/odata/Objects"
	query := fmt.Sprintf(
		`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('Owner'))&$filter=Model/Name eq 'Baseline Architecture' and ObjectType/Name in ('Physical Application Component','Logical Application Component','Physical Technology Component') and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'GU::Domain' and a/Values/any(b:indexof(b/Value,'%s') gt -1)) and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/all(b:indexof(b/Value,'Retired') eq -1 and indexof(b/Value,'Proposed') eq -1))`,
		department[1:6],
	)
	query = strings.ReplaceAll(query, " ", "%20")
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, x := range oneCall.Value {
			toReturn[x.AttributeValues[0].StringValue] = append(toReturn[x.AttributeValues[0].StringValue], x.ObjectId)
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}

	putInto(toReturn, thenWindow)
}

func (a *AzureAuth) GetDomainThen(department string, putInto laterDomainOwned, thenWindow fyne.Window) {
	toReturn := map[string][]IServerObjectStruct{}

	type objects struct {
		Value    []IServerObjectStruct `json:"value"`
		NextLink string                `json:"@odata.nextLink"`
	}

	path := "/odata/Objects"
	query := fmt.Sprintf(
		`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName eq 'Owner')&$filter=Model/Name eq 'Baseline Architecture' and ObjectType/Name in ('Physical Application Component','Physical Technology Component') and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'GU::Domain' and a/Values/any(b:b/Value eq '%s')) and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/all(b:indexof(b/Value,'Retired') eq -1 and indexof(b/Value,'Proposed') eq -1))`,
		strings.ReplaceAll(department, "&", "%26"),
	)
	query = strings.ReplaceAll(query, " ", "%20")
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, x := range oneCall.Value {
			dept := x.AttributeValues[0].StringValue
			if len(dept) == 0 {
				dept = "???"
			}
			x.AttributeValues = []AttributeValue{}
			toReturn[dept] = append(
				toReturn[dept],
				x)
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	putInto(toReturn, thenWindow)
}

func (a *AzureAuth) GetChoicesFor(me string) map[string]string {
	var oneCall struct {
		Choices []struct {
			Value                          string `json:"Value"`
			AttributeConfigurationChoiceId string `json:"AttributeConfigurationChoiceId"`
		}
		NextLink string `json:"@odata.nextLink"`
	}
	// Lifecycle
	Choices := map[string]string{}
	path := fmt.Sprintf("/odata/Attributes(%s)", me)
	query := ""
	for {
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, x := range oneCall.Choices {
			Choices[x.Value] = x.AttributeConfigurationChoiceId
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	return Choices
}

func (a *AzureAuth) GetChoicesForName(me string) map[string]string {
	var oneCall struct {
		Value []struct {
			Choices []struct {
				Value                          string `json:"Value"`
				AttributeConfigurationChoiceId string `json:"AttributeConfigurationChoiceId"`
			} `json:"Choices"`
		} `json:"value"`
		NextLink string `json:"@odata.nextLink"`
	}
	// Lifecycle
	Choices := map[string]string{}
	path := "/odata/Attributes"
	query := `$filter=` + url.QueryEscape(fmt.Sprintf("Name eq '%s'", me))
	for {
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		err = json.Unmarshal(bytemep, &oneCall)
		if err != nil {
			log.Fatalf("failed to parse json %v\n", err)
		}
		if len(oneCall.Value) > 0 && len(oneCall.Value[0].Choices) > 0 {
			for _, x := range oneCall.Value[0].Choices {
				Choices[x.Value] = x.AttributeConfigurationChoiceId
			}
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	return Choices

}

// Simple find over iServer components, looking for the specified string
func (a *AzureAuth) FindMeInTypeThen(
	lookFor string,
	objectType string,
	putInto func([]FindStruct)) {

	toReturn := []FindStruct{}

	type objects struct {
		Value    []FindStruct `json:"value"`
		NextLink string       `json:"@odata.nextLink"`
	}

	path := "/odata/Objects"
	query := strings.ReplaceAll(
		fmt.Sprintf(
			`$expand=AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('ObjectId','Name','ObjectType','GU::Level'))&$filter=Model/Name eq 'Baseline Architecture' and ObjectType/ObjectTypeId eq %s and indexOf(tolower(Name),'%s') gt -1`,
			objectType,
			strings.ToLower(lookFor)),
		" ",
		"%20")
	if objectType == "265f5bb2-2eef-e811-9f2b-00155d26bcf8" {
		query = query + strings.ReplaceAll("and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueNumber/any(a:a/AttributeName eq 'GU::Level' and a/Value eq 2)", " ", "%20")
	}
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		toReturn = append(toReturn, oneCall.Value...)
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}

	putInto(toReturn)
}

// EXCEL FUNCTIONS

func (a *AzureAuth) GetRelationsAsSliceString(objectid, objecttype string) map[string][]string {
	returns := map[string][]string{
		"Capabilities": {},
	}
	for _, x := range a.FindRelations(objectid) {
		target := x.MemberObject
		if x.MemberObjectId == objectid {
			target = x.LeadObject
		}
		returns[target.Type.Name] = append(returns[target.Type.Name], target.Name)
	}
	return returns
}

// Create the excel ProductManager overview report from iserver data
func (a *AzureAuth) CreateProductManagerOverviewReport(department, savePath string) {
	f := excelize.NewFile()

	// Header style
	style_header, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: "000000", Style: 3},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 20,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	rowidx := 1
	row := []interface{}{
		"Object ID",
		"Object Name",
		"Object Type",
		"Product Manager",
		"Business Owner",
		"Serviceability",
		"Lifecycle",
		"Capabilities",
		"Physical Application Component",
		"Physical Technology Component",
		"Physical Data Component",
	}
	cell, err := excelize.CoordinatesToCellName(1, rowidx)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetSheetRow("Sheet1", cell, &row)
	f.SetCellStyle("Sheet1", "A1", "K1", style_header)
	f.SetColWidth("Sheet1", "B", "B", 62)
	f.SetColWidth("Sheet1", "D", "D", 41.33)
	f.SetColWidth("Sheet1", "E", "E", 62.17)
	f.SetColWidth("Sheet1", "F", "F", 133.5)
	f.SetColWidth("Sheet1", "G", "G", 17.17)
	f.SetColWidth("Sheet1", "H", "H", 31.17)
	f.SetColWidth("Sheet1", "I", "K", 52.17)
	f.SetColVisible("Sheet1", "A", false)
	f.SetColVisible("Sheet1", "C", false)

	// Wrap style
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	type objects struct {
		Value    []IServerObjectStruct `json:"value"`
		NextLink string                `json:"@odata.nextLink"`
	}

	// Get all PAC and PTC by Product Manager
	path := "/odata/Objects"
	query := fmt.Sprintf(
		`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('Owner','Lifecycle Status','Serviceability characteristics','Department'))&$filter=Model/Name eq 'Baseline Architecture' and ObjectType/Name in ('Physical Application Component','Physical Technology Component') and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'GU::Domain' and a/Values/any(b:indexof(b/Value,'%s') gt -1)) and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/all(b:indexof(b/Value,'Retired') eq -1))`,
		department[1:6],
	)
	query = strings.ReplaceAll(query, " ", "%20")
	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		defer mep.Close()
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, x := range oneCall.Value {
			fieldmap := map[string]interface{}{}
			for _, y := range x.AttributeValues {
				fieldmap[y.AttributeName] = y.StringValue
			}
			// Get all related Capabilities, PTC/PAC, Data items
			rels := a.GetRelationsAsSliceString(x.ObjectId, x.ObjectType.Id)
			// Cell
			rowidx = rowidx + 1
			row := []interface{}{
				x.ObjectId,
				x.Name,
				x.ObjectType.Name,
				fieldmap["Owner"],
				fieldmap["Department"],
				fieldmap["Serviceability characteristics"],
				fieldmap["Lifecycle Status"],
				strings.Join(rels["Capability"], "\r\n"),
				strings.Join(rels["Physical Application Component"], "\r\n"),
				strings.Join(rels["Physical Technology Component"], "\r\n"),
				strings.Join(rels["Physical Data Component"], "\r\n"),
			}
			cell, err := excelize.CoordinatesToCellName(1, rowidx)
			if err != nil {
				fmt.Println(err)
				return
			}
			f.SetSheetRow("Sheet1", cell, &row)
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			log.Printf("Failed to parse next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	cell, _ = excelize.CoordinatesToCellName(11, rowidx)
	f.SetCellStyle("Sheet1", "H2", cell, style)
	f.AddTable("Sheet1", &excelize.Table{Range: "A1:" + cell})
	// Export as an Excel report
	if err := f.SaveAs(filepath.Join(savePath, "iServerAudit.xlsx")); err != nil {
		fmt.Println(err)
	}
}
