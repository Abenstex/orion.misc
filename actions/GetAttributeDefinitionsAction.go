package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	utils2 "laniakea/utils"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	structs2 "orion.commons/structs"
	"orion.commons/utils"
	"orion.misc/structs"
	"time"
)

const SqlGetAllAttributeDefinitions = "SELECT id, name, description, active, " +
	"(extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, b.action_by, datatype, overwriteable, " +
	"allowed_object_types, list_of_values, numeric_from, numeric_to, query, default_value, assign_during_object_creation " +
	"FROM attributes a left outer join cache b on a.id=b.object_id "

type GetAttributeDefinitionsAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetAttributeDefinitionsAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetStatesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetAttributeDefinitionsAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetAttributeDefinitionsAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetAttributeDefinitionsAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetAttributeDefinitionsAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetAttributeDefinitionsAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetAttributeDefinitionsAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetAttributeDefinitionsAction) SendEvents(request micro.IRequest) {

}

func (action GetAttributeDefinitionsAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/get"
	var error = "orion/server/misc/error/attributedefinition/get"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "GetAttributeDefinitionsAction",
		Description:    "Get attribute definitions based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/attributedefinition/get",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.baseAction.ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *GetAttributeDefinitionsAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetAttributeDefinitionsAction) createGetAttributeDefinitionsReply(definitions []structs2.AttributeDefinition) (structs.GetAttributeDefinitionsReply, *micro.Exception) {
	var reply = structs.GetAttributeDefinitionsReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(definitions) > 0 {
		reply.Header.Success = true
		reply.AttributeDefinitions = definitions
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No definitions were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetAttributeDefinitionsAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetAttributeDefinitionsRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	reply, myErr := action.getAttributeDefinitions(receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetAttributeDefinitionsAction) fillAttributeDefinitions(rows *sql.Rows) ([]structs2.AttributeDefinition, *micro.Exception) {
	var attributes []structs2.AttributeDefinition

	for rows.Next() {
		var attribute structs2.AttributeDefinition
		/*
			id, name, description, active, " +
			"(extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, b.action_by, datatype, overwriteable, " +
			"allowed_object_types, list_of_values, numeric_from, numeric_to, query, default_value, assign_during_object_creation
		*/
		err := rows.Scan(&attribute.Info.Id, &attribute.Info.Name, &attribute.Info.Description, &attribute.Info.Active,
			&attribute.Info.CreatedDate, &attribute.Info.Alias, &attribute.Info.LockedBy, &attribute.DataType,
			&attribute.Overwriteable, pq.Array(&attribute.AllowedObjectTypes), pq.Array(&attribute.ListOfValues),
			&attribute.NumericFrom, &attribute.NumericTo, &attribute.Query, &attribute.DefaultValue, &attribute.AssignDuringObjectCreation)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		attributes = append(attributes, attribute)
	}
	//fmt.Sprintf("Size of attributes: %d", len(attributes))
	return attributes, nil
}

func (action GetAttributeDefinitionsAction) getAttributeDefinitions(request structs.GetAttributeDefinitionsRequest) (structs.GetAttributeDefinitionsReply, *micro.Exception) {
	attributes, myErr := action.getAttributeDefinitionsFromDb(request)

	if myErr != nil {
		return structs.GetAttributeDefinitionsReply{}, myErr
	}

	return action.createGetAttributeDefinitionsReply(attributes)
}

func (action GetAttributeDefinitionsAction) getAttributeDefinitionsFromDb(request structs.GetAttributeDefinitionsRequest) ([]structs2.AttributeDefinition, *micro.Exception) {
	var sql = SqlGetAllAttributeDefinitions

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetAttributeDefinitionsAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetAttributeDefinitionsAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	states, myErr := action.fillAttributeDefinitions(rows)
	if myErr != nil {
		return states, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return states, nil
}
