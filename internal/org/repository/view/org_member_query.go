package view

import (
	global_model "github.com/caos/zitadel/internal/model"
	proj_model "github.com/caos/zitadel/internal/org/model"
	"github.com/caos/zitadel/internal/view"
)

type OrgMemberSearchRequest proj_model.OrgMemberSearchRequest
type OrgMemberSearchQuery proj_model.OrgMemberSearchQuery
type OrgMemberSearchKey proj_model.OrgMemberSearchKey

func (req OrgMemberSearchRequest) GetLimit() uint64 {
	return req.Limit
}

func (req OrgMemberSearchRequest) GetOffset() uint64 {
	return req.Offset
}

func (req OrgMemberSearchRequest) GetSortingColumn() view.ColumnKey {
	if req.SortingColumn == proj_model.ORGMEMBERSEARCHKEY_UNSPECIFIED {
		return nil
	}
	return OrgMemberSearchKey(req.SortingColumn)
}

func (req OrgMemberSearchRequest) GetAsc() bool {
	return req.Asc
}

func (req OrgMemberSearchRequest) GetQueries() []view.SearchQuery {
	result := make([]view.SearchQuery, len(req.Queries))
	for i, q := range req.Queries {
		result[i] = OrgMemberSearchQuery{Key: q.Key, Value: q.Value, Method: q.Method}
	}
	return result
}

func (req OrgMemberSearchQuery) GetKey() view.ColumnKey {
	return OrgMemberSearchKey(req.Key)
}

func (req OrgMemberSearchQuery) GetMethod() global_model.SearchMethod {
	return req.Method
}

func (req OrgMemberSearchQuery) GetValue() interface{} {
	return req.Value
}

func (key OrgMemberSearchKey) ToColumnName() string {
	switch proj_model.OrgMemberSearchKey(key) {
	case proj_model.ORGMEMBERSEARCHKEY_EMAIL:
		return OrgMemberKeyEmail
	case proj_model.ORGMEMBERSEARCHKEY_FIRST_NAME:
		return OrgMemberKeyFirstName
	case proj_model.ORGMEMBERSEARCHKEY_LAST_NAME:
		return OrgMemberKeyLastName
	case proj_model.ORGMEMBERSEARCHKEY_USER_NAME:
		return OrgMemberKeyUserName
	case proj_model.ORGMEMBERSEARCHKEY_USER_ID:
		return OrgMemberKeyUserID
	case proj_model.ORGMEMBERSEARCHKEY_ORG_ID:
		return OrgMemberKeyOrgID
	default:
		return ""
	}
}