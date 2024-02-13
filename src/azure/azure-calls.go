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
			//owners[i] = strings.Replace(owner, "&", "\\u0026", -1)
			owners = append(owners, strings.Replace(owner, "&", "and", -1))
		} else if strings.Contains(owner, "and") {
			owners = append(owners, strings.Replace(owner, "and", "&", -1))
		}
	}
	fmt.Printf("Match against owners %v\n", owners)
	filterQuery = fmt.Sprintf(
		` and AttributeValues/OfficeArchitect.Contracts.OData.Model.AttributeValue.AttributeValueChoice/any(a:a/AttributeName eq '%s' and a/Values/any(b:b/Value eq '%s'))`+
			` and AttributeValues/OfficeArchitect.Contracts.OData.Model.Attributevalue.AttributeValueChoice/any(a:a/AttributeName eq 'Lifecycle Status' and a/Values/any(b:b/value in ('Proposed','In Development','Live','Phasing Out')))`,
		"GU::Domain",
		strings.Replace(department, "&", "%26", -1),
	)
	path := "/odata/Objects"
	query := fmt.Sprintf(`$expand=ObjectType($select=Name),AttributeValues($select=StringValue,AttributeName;$filter=AttributeName in ('Lifecycle Status','IServerID','Description','Owner','GU::Review Bodies','Owner (Legacy)','Internal Recommendation','Operational Importance'))&`+
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
				fmt.Printf("Adding %s\n", anObject.Name)
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
	return toReturn
}
