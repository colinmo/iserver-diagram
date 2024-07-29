package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
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

func (a *AzureAuth) GetImportantFields(id string) IServerObjectStruct {
	toReturn := IServerObjectStruct{}

	path := fmt.Sprintf("/odata/Objects(%s)", id)
	query := `$expand=ObjectType($select=Name,ObjectTypeId),AttributeValues($select=StringValue,AttributeName,AttributeId;$filter=AttributeName%20in%20('Description','Owner','Owner%20(Legacy)','GU::Managed%20Outside%20Of%20DS','GU::Information%20System%20Custodian','GU::Review%20Bodies','Lifecycle%20Status','GU::Information%20Security%20Classification','GU::Object%20Visibility','GU::Solution%20Classification','Internal:%20In%20Development%20From','Internal:%20Live%20Date','Internal:%20Phase%20Out%20From','Internal:%20Retirement%20Date','Supplier','Internal%20Recommendation','Operational%20Importance','GU::Domain','Department'))`
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
	saveValues.AttributeValues = append(saveValues.AttributeValues, SaveValue{
		AttributeName:     "Name",
		AttributeCategory: "Text",
		TextValue:         stringValues["Title"],
	})
	delete(stringValues, "Title")
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
		return toReturn.Success, toReturn.Messages[0].Message, toReturn.SuccessMessage.MessageDefinition.ObjectId
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

func (a *AzureAuth) FindRelationsThen(id string, putInto laterRelationUpdate, thenWindow *fyne.Window) {
	putInto(a.GetImportantFields(id), a.FindRelations(id), thenWindow)
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
	query := strings.ReplaceAll(fmt.Sprintf("%%24filter=Name eq '%s'", me), " ", "%20")
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
