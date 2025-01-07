package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"
)

/*
	Learnings

	* When using `$expand` in an object query, the values aren't hydrated!
		* the only way to get the values is to do GetAttributes calls
*/

var defaultModel = "Baseline Architecture"

var objectTypesList = map[string]string{
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
var ObjectTypesListLookup = map[string]string{
	"445f5bb2-2eef-e811-9f2b-00155d26bcf8": "Actor",
	"bc5f5bb2-2eef-e811-9f2b-00155d26bcf8": "Application Service",
	"73d7af8c-5e52-ea11-a94c-28187852a561": "Business Service",
	"265f5bb2-2eef-e811-9f2b-00155d26bcf8": "Capability",
	"535f5bb2-2eef-e811-9f2b-00155d26bcf8": "Constraint",
	"625f5bb2-2eef-e811-9f2b-00155d26bcf8": "Data Entity",
	"a3b624e4-b642-ea11-a601-28187852aafd": "Interface",
	"cb5f5bb2-2eef-e811-9f2b-00155d26bcf8": "Location",
	"7cb624e4-b642-ea11-a601-28187852aafd": "Logical Application Component",
	"96b624e4-b642-ea11-a601-28187852aafd": "Logical Data Component",
	"070714ec-b642-ea11-a601-28187852aafd": "Logical Technology Component",
	"b0b624e4-b642-ea11-a601-28187852aafd": "Organization Unit",
	"6fb624e4-b642-ea11-a601-28187852aafd": "Physical Application Component",
	"37395db8-2eef-e811-9f2b-00155d26bcf8": "Physical Data Component",
	"140714ec-b642-ea11-a601-28187852aafd": "Physical Technology Component",
	"5171e716-436d-ee11-9942-00224895c2e5": "Physical Technology Group",
	"70395db8-2eef-e811-9f2b-00155d26bcf8": "Principle",
	"7f395db8-2eef-e811-9f2b-00155d26bcf8": "Process",
	"8e395db8-2eef-e811-9f2b-00155d26bcf8": "Product",
	"9d395db8-2eef-e811-9f2b-00155d26bcf8": "Requirement",
	"b65f6dbe-2eef-e811-9f2b-00155d26bcf8": "Risk",
	"243a5db8-2eef-e811-9f2b-00155d26bcf8": "Role",
	"52395db8-2eef-e811-9f2b-00155d26bcf8": "Technology Service",
}

func (a *AzureAuth) WhoAmI() {
	mep, err := a.CallRestEndpoint("GET", "/odata/Me", []byte{}, "")
	if err != nil {
		log.Fatalf("failed to call endpoint %v\n", err)
	}
	bytemep, err := io.ReadAll(mep)
	if err != nil {
		log.Fatalf("failed to read io.Reader %v\n", err)
	}
	fmt.Printf("%s\n\n%v", string(bytemep), err)
}

type ObjectTypeStruct struct {
	Color              string `json:"Color"`
	Icon               string `json:"Icon"`
	ObjectTypeId       string `json:"ObjectTypeId"`
	ParentObjectTypeId string `json:"ParentObjectTypeId"`
	DateCreated        string `json:"DateCreated"`
	CreatedById        string `json:"CreatedById"`
	DateLastModified   string `json:"DateLastModified"`
	LastModifiedById   string `json:"LastModifiedById"`
	ActiveState        bool   `json:"ActiveState"`
	IsApprovable       bool   `json:"IsApprovable"`
	Alias              string `json:"Alias"`
	Name               string `json:"Name"`
	Description        string `json:"Description"`
}

type AttributeTypeStruct struct {
	StringValue         string   `json:"StringValue"`
	AttributeValueId    int32    `json:"AttributeValueId"`
	AttributeCategoryId int32    `json:"AttributeCategoryId"`
	AttributeId         string   `json:"AttributeId"`
	AttributeName       string   `json:"AttributeName"`
	Value               string   `json:"Value"`
	Values              []string `json:"Values"`
}

type ObjectStruct struct {
	ObjectID         string                `json:"ObjectId"`
	Name             string                `json:"Name"`
	ObjectTypeId     string                `json:"ObjectTypeId"`
	LockedOn         bool                  `json:"LockedOn"`
	LockedById       bool                  `json:"LockedById"`
	IsApproved       bool                  `json:"IsApproved"`
	ModelId          string                `json:"ModelId"`
	DateCreated      string                `json:"DateCreated"`
	CreatedById      string                `json:"CreatedById"`
	LastModifiedDate string                `json:"LastModifiedDate"`
	LastModifiedById string                `json:"LastModifiedById"`
	Attributevalues  []AttributeTypeStruct `json:"AttributeValues"`
	ObjectType       ObjectTypeStruct      `json:"ObjectType"`
}

type RelationshipStruct struct {
	RelationshipId         string `json:"RelationshipId"`
	RelationshipTypeId     string `json:"RelationshipTypeId"`
	LeadObjectId           string `json:"LeadObjectId"`
	MemberObjectId         string `json:"MemberObjectId"`
	RelationshipTypePairId string `json:"RelationshipTypePairId"`
	IsApproved             bool   `json:"IsApproved"`
	ModelId                string `json:"ModelId"`
	DateCreated            string `json:"DateCreated"`
	CreatedById            string `json:"CreatedById"`
	LastModifiedDate       string `json:"LastModifiedDate"`
	LastModifiedById       string `json:"LastModifiedById"`
	RelationshipType       struct {
		Color                 string `json:"Color"`
		RelationshipTypeId    string `json:"RelationshipTypeId"`
		DateCreated           string `json:"DateCreated"`
		CreatedById           string `json:"CreatedById"`
		DateLastModified      string `json:"DateLastModified"`
		LastModifiedById      string `json:"LastModifiedById"`
		ActiveState           bool   `json:"ActiveState"`
		IsApprovable          bool   `json:"IsApprovable"`
		LeadToMemberDirection string `json:"LeadToMemberDirection"`
		MemberToLeadDirection string `json:"MemberToLeadDirection"`
		Representation        string `json:"Representation"`
		Direction             string `json:"Direction"`
		Alias                 string `json:"Alias"`
		Name                  string `json:"Name"`
		Description           string `json:"Description"`
	} `json:"RelationshipType"`
	LeadObject   ObjectTypeStruct `json:"LeadObject"`
	MemberObject ObjectTypeStruct `json:"MemberObject"`
}

type MinRelationship struct {
	RelationshipID        string
	RelationshipType      string
	LeadObjectID          string
	LeadObjectName        string
	MemberObjectID        string
	MemberObjectName      string
	LeadToMemberDirection string
}

type RelationshipTypeStruct struct {
	RelationshipTypeId    string `json:"RelationshipTypeId"`
	ActiveState           bool   `json:"ActiveState"`
	Direction             string `json:"Direction"`
	Name                  string `json:"Name"`
	RelationshipTypePairs []struct {
		RelationshipTypePairId string `json:"RelationshipTypePairId"`
		LeadObjectTypeId       string `json:"LeadObjectTypeId"`
	} `json:"RelationshipTypePairs,omitempty"`
}

func (a *AzureAuth) GetObjectsByCategory(category string, attributes []string) map[string]ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := map[string]ObjectStruct{}
	path := "/odata/Objects"
	query := `$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName%20eq%20'Lifecycle%20Status')&$filter=Model/Name%20eq%20'` + url.QueryEscape(defaultModel) + `'%20and%20AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueText/any(a:a/AttributeName%20eq%20'Category%20(General)'%20and%20a/Value%20eq%20'` + url.QueryEscape(category) + `')`

	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, x := range oneCall.Value {
			toReturn[x.ObjectID] = x
		}
		if len(oneCall.NextLink) == 0 {
			break
		}
		bits, err := url.Parse(oneCall.NextLink)
		if err != nil {
			fmt.Printf("Failed to get next")
			break
		}
		path = bits.Path
		query = bits.RawQuery
		time.Sleep(100 * time.Millisecond)
	}
	return toReturn
}

func (a *AzureAuth) GetAllObjects(attributes []string) []ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	attributeQuery := ""
	if len(attributes) > 0 {
		attributeQuery = fmt.Sprintf(";$filter=AttributeName eq '%s'", strings.Join(attributes, "' or AttributeName eq '"))
		attributeQuery = strings.ReplaceAll(attributeQuery, " ", "%20")
	}

	path := "/odata/Objects"
	query := `$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName` + attributeQuery + `)&$filter=Model/Name%20eq%20'` + url.QueryEscape(defaultModel) + `'`

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

func (a *AzureAuth) GetAllObjectsOfType(objectType string, attributes []string) []ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	attributeQuery := ""
	if len(attributes) > 0 {
		attributeQuery = fmt.Sprintf(";$filter=AttributeName eq '%s'", strings.Join(attributes, "' or AttributeName eq '"))
		attributeQuery = strings.ReplaceAll(attributeQuery, " ", "%20")
	}

	path := "/odata/Objects"
	query := `$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName` + attributeQuery + `)&$filter=Model/Name%20eq%20'` + url.QueryEscape(defaultModel) + `'%20and%20ObjectType/Name%20eq%20'` + strings.ReplaceAll(objectType, " ", "%20") + `'`

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

func (a *AzureAuth) GetLeadRelationshipsForObject(objectId string) map[string]MinRelationship {
	type objects struct {
		Value    []RelationshipStruct `json:"value"`
		NextLink string               `json:"@odata.nextLink"`
	}
	toReturn := map[string]MinRelationship{}
	path := "/odata/Relationships"
	query := fmt.Sprintf(`$expand=RelationshipType,LeadObject,MemberObject&$filter=LeadObjectId%%20eq%%20%s`, objectId)

	for {
		var oneCall objects
		mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
		if err != nil {
			log.Fatalf("failed to call endpoint %v\n", err)
		}
		bytemep, err := io.ReadAll(mep)
		json.Unmarshal(bytemep, &oneCall)

		if err != nil {
			log.Fatalf("failed to read io.Reader %v\n", err)
		}
		for _, rel := range oneCall.Value {
			toReturn[rel.RelationshipId] = MinRelationship{
				RelationshipID:        rel.RelationshipId,
				RelationshipType:      rel.RelationshipType.RelationshipTypeId,
				LeadObjectID:          rel.LeadObjectId,
				LeadObjectName:        rel.LeadObject.Name,
				MemberObjectID:        rel.MemberObjectId,
				MemberObjectName:      rel.MemberObject.Name,
				LeadToMemberDirection: rel.RelationshipType.LeadToMemberDirection,
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
	return toReturn
}

func (a *AzureAuth) GetObjectsForTypeAndArea(objectType string, owners []string) []ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	filterQuery := ""
	for i, owner := range owners {
		if strings.Contains(owner, "&") {
			owners[i] = strings.Replace(owner, "&", "%26", -1)
			owners = append(owners, strings.Replace(owner, "&", "and", -1))
		}
	}
	owner := strings.Join(owners, "','")
	switch objectType {
	case "PAC":
		objectType = "Physical Application Component"
		filterQuery = fmt.Sprintf(
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueText/any(a:`+
				`a/AttributeName eq '%s' and a/Value in ('%s')`+
				`)`,
			"Owner",
			owner,
		)
	case "PTC":
		objectType = "Physical Technology Component"
		filterQuery = fmt.Sprintf(
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueText/any(a:`+
				`a/AttributeName eq '%s' and a/Value in ('%s')`+
				`)`,
			"Owner",
			owner,
		)
	case "LAC":
		objectType = "Logical Application Component"
		filterQuery = fmt.Sprintf(
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueText/any(a:`+
				`a/AttributeName eq '%s' and a/Value in ('%s')`+
				`)`,
			"Owner",
			owner,
		)
	}
	path := "/odata/Objects"
	query := fmt.Sprintf(`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('Business Fit','Technical Fit','Lifecycle Status','IServerID','Internal: In Development From','Internal: Live date','Internal: Phase Out From','Internal: Retirement date','Internal Recommendation','Operational Importance'))&`+
		`$filter=Model/Name eq '%s'`+
		` and ObjectType/Name eq '%s'`+
		`%s`,
		defaultModel,
		objectType,
		filterQuery)
	query = strings.Replace(query, " ", "%20", -1)
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

func (a *AzureAuth) GetObjectsForTypeAndDepartmentWithoutOwners(objectType string, department string, owners []string) []ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	filterQuery := ""
	bob := owners
	for _, owner := range bob {
		if strings.Contains(owner, "&") {
			owners = append(owners, strings.Replace(owner, "&", "and", -1))
		} else if strings.Contains(owner, "and") {
			owners = append(owners, strings.Replace(owner, "and", "&", -1))
		}
	}
	filterQuery = fmt.Sprintf(
		` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq '%s' and a/Values/any(b:b/Value eq '%s'))`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.Attributevalue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/any(b:b/value in ('Proposed','In Development','Live','Phasing Out')))`,
		"GU::Domain",
		strings.Replace(department, "&", "%26", -1),
	)
	path := "/odata/Objects"
	query := fmt.Sprintf(`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('Lifecycle Status','IServerID','Description','Owner','GU::Review Bodies','Owner (Legacy)','Internal Recommendation','Operational Importance','Department'))&`+
		`$filter=Model/Name eq '%s'`+
		` and ObjectType/Name in ('Physical Application Component','Physical Technology Component','Logical Application Component')`+
		`%s`,
		defaultModel,
		filterQuery)
	query = strings.Replace(query, " ", "%20", -1)
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
		for _, anObject := range oneCall.Value {
			if unknownProductManager(anObject, owners) {
				toReturn = append(toReturn, anObject)
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
	return toReturn
}

func unknownProductManager(needle ObjectStruct, haystack []string) bool {
	for _, a := range needle.Attributevalues {
		if a.AttributeName == "Owner" {
			for _, v := range haystack {
				if a.StringValue == v {
					//fmt.Printf("Do not add [%s] vs [%s]\n", a.StringValue, v)
					return false
				}
			}
			fmt.Printf("Add [%s]\n", a.StringValue)
		}
	}
	return true
}

func (a *AzureAuth) GetRelationTypesForObjectType(objectTypeId1, objectTypeId2 string) map[string]RelationshipTypeStruct {
	type objects struct {
		Value    []RelationshipTypeStruct `json:"value"`
		NextLink string                   `json:"@odata.nextLink"`
	}
	toReturn := map[string]RelationshipTypeStruct{}
	path := fmt.Sprintf("/odata/RelationshipTypes/GetByObjectTypes(objectTypeId1=%s,objectTypeId2=%s)", objectTypeId1, objectTypeId2)
	query := `$expand=RelationshipTypePairs`

	if objectTypeId1 != "" && objectTypeId2 != "" {
		for {
			var oneCall objects
			mep, err := a.CallRestEndpoint("GET", path, []byte{}, query)
			if err != nil {
				log.Fatalf("failed to call endpoint %v\n", err)
			}
			bytemep, err := io.ReadAll(mep)
			json.Unmarshal(bytemep, &oneCall)

			if err != nil {
				log.Fatalf("failed to read io.Reader %v\n", err)
			}
			for _, rel := range oneCall.Value {
				toReturn[rel.RelationshipTypeId] = rel
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
	}
	return toReturn
}

func (a *AzureAuth) DeleteARelationship(id string) error {
	var err error
	var mep io.ReadCloser
	path := fmt.Sprintf("/odata/Relationships(%s)", id)
	mep, err = a.CallRestEndpoint("DELETE", path, []byte{}, "")
	if err == nil {
		defer mep.Close()
		var bytemep []byte
		var messageResponse ODataResponse
		bytemep, err = io.ReadAll(mep)
		json.Unmarshal(bytemep, &messageResponse)
		if !messageResponse.Success {
			return fmt.Errorf("%s", messageResponse.SuccessMessage.MessageCode)
		}
		return err
	}
	return err
}

type ODataMessage struct {
	MessageCategory   string `json:"messageCategory"`
	MessageCode       string `json:"messageCode"`
	MessageDefinition struct {
		DeletedRelationshipID string `json:"deletedRelationshipId"`
	} `json:"messageDefinition"`
}
type ODataResponse struct {
	OperationType  string         `json:"operationType"`
	EntityTypes    string         `json:"entityTypes"`
	SuccessMessage ODataMessage   `json:"successMessage"`
	Success        bool           `json:"success"`
	Messages       []ODataMessage `json:"messages"`
}

// 2025

func (a *AzureAuth) GetPACForRSDFDomain() []ObjectStruct {
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	path := "/odata/Objects"
	query := fmt.Sprintf(
		`$filter=Model/Name eq '%s'`+
			` and ObjectType/Name eq 'Physical Application Component'`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'GU::Domain' and a/Values/any(b:b/Value eq 'Research, Specialised & Data Foundations'))`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/any(b:b/Value in ('In Development','Live')))`,
		defaultModel)
	query = strings.Replace(query, " ", "%20", -1)
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

func (a *AzureAuth) GetDomainObjectsForHERM() []ObjectStruct {
	// * PAC - Our specific applications
	type objects struct {
		Value    []ObjectStruct `json:"value"`
		NextLink string         `json:"@odata.nextLink"`
	}
	toReturn := []ObjectStruct{}
	path := "/odata/Objects"
	query := fmt.Sprintf(
		`$filter=Model/Name eq '%s'`+
			` and ObjectType/Name eq 'Physical Application Component'`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'GU::Domain' and a/Values/any(b:b/Value eq 'Research, Specialised %%26 Data Foundations'))`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/any(b:b/Value in ('In Development','Live')))`,
		defaultModel)
	query = strings.Replace(query, " ", "%20", -1)
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

func (a *AzureAuth) GetRelatedHERMObjects(objectsin []ObjectStruct) ([]ObjectStruct, []MinRelationship) {
	// PAC links to PAC, LAC, PDC, PTC, CAP
	// * CAP - BCM
	// * LAC - ARM
	// LAC links to LTC @todo
	// * LTC - TRM
	// PDC links to LDC @todo
	// * LDC - DRM
	type objects struct {
		Value    []RelationshipStruct `json:"value"`
		NextLink string               `json:"@odata.nextLink"`
	}
	toReturnObjects := []ObjectStruct{}
	uniqueObjects := map[string]ObjectStruct{}
	toReturnRelations := []MinRelationship{}
	uniqueRelations := map[string]RelationshipStruct{}
	// Convert objectsin to just ids
	objectIds := []string{}
	for _, x := range objectsin {
		objectIds = append(objectIds, x.ObjectID)
	}
	relatedObjects := []string{
		objectTypesList["Logical Application Component"],
		objectTypesList["Physical Data Component"],
		objectTypesList["Physical Technology Component"],
	}

	path := "/odata/Relationships"
	for _, query := range []string{
		fmt.Sprintf(
			`$expand=LeadObject($select=Name,ObjectTypeId),MemberObject($select=Name,ObjectTypeId)&`+
				`$filter=Model/Name eq '%s'`+
				` and LeadObjectId in (%s) and MemberObject/ObjectTypeId in (%s)`,
			defaultModel,
			strings.Join(objectIds, ","),
			strings.Join(relatedObjects, ","),
		),
		fmt.Sprintf(
			`$expand=LeadObject($select=Name,ObjectTypeId),MemberObject($select=Name,ObjectTypeId)&`+
				`$filter=Model/Name eq '%s'`+
				` and MemberObjectId in (%s) and LeadObject/ObjectTypeId in (%s)`,
			defaultModel,
			strings.Join(objectIds, ","),
			strings.Join(relatedObjects, ","),
		),
		fmt.Sprintf(
			`$expand=LeadObject($select=Name,ObjectTypeId),MemberObject($select=Name,ObjectTypeId)&`+
				`$filter=Model/Name eq '%s'`+
				` and LeadObjectId in (%s) and MemberObject/ObjectTypeId in (%s)`,
			defaultModel,
			strings.Join(objectIds, ","),
			objectTypesList["Capability"],
		),
		fmt.Sprintf(
			`$expand=LeadObject($select=Name,ObjectTypeId),MemberObject($select=Name,ObjectTypeId)&`+
				`$filter=Model/Name eq '%s'`+
				` and MemberObjectId in (%s) and LeadObject/ObjectTypeId in (%s)`,
			defaultModel,
			strings.Join(objectIds, ","),
			objectTypesList["Capability"],
		),
	} {
		query = strings.Replace(query, " ", "%20", -1)
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
				uniqueRelations[x.RelationshipId] = x
				uniqueObjects[x.LeadObjectId] = ObjectStruct{ObjectID: x.LeadObjectId, Name: x.LeadObject.Name, ObjectType: ObjectTypeStruct{Name: ObjectTypesListLookup[x.LeadObject.ObjectTypeId]}}
				uniqueObjects[x.MemberObjectId] = ObjectStruct{ObjectID: x.MemberObjectId, Name: x.MemberObject.Name, ObjectType: ObjectTypeStruct{Name: ObjectTypesListLookup[x.MemberObject.ObjectTypeId]}}
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
			time.Sleep(200 * time.Millisecond)
		}
	}
	for _, x := range uniqueRelations {
		toReturnRelations = append(toReturnRelations, MinRelationship{LeadObjectID: x.LeadObjectId, MemberObjectID: x.MemberObjectId, RelationshipType: ObjectTypesListLookup[x.RelationshipTypeId]})
	}
	for _, x := range uniqueObjects {
		toReturnObjects = append(toReturnObjects, x)
	}
	return toReturnObjects, toReturnRelations
}
