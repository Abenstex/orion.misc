package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

const SqlGetAttributeValues = "SELECT attr_id, attr_value, a.object_type, object_id, b.datatype, b.name, b.description, " +
	"(extract(epoch from b.created_date)::bigint)*1000 AS created_date, " +
	"b.active, b.pretty_id, a.id, b.overwriteable " +
	"FROM ref_attributes_objects a INNER JOIN attributes b ON a.attr_id = b.id " +
	"WHERE a.object_id=$1 "

type GetAttributeValuesPerObjectAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetAttributeValuesPerObjectAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetAttributeValuesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	if dummy.ObjectId == nil {
		return micro.NewException(structs2.MissingParameterError, fmt.Errorf("the parameter objectId is missing in the request"))
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetAttributeValuesPerObjectAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetAttributeValuesPerObjectAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetAttributeValuesPerObjectAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetAttributeValuesPerObjectAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetAttributeValuesPerObjectAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetAttributeValuesPerObjectAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetAttributeValuesPerObjectAction) SendEvents(request micro.IRequest) {

}

func (action GetAttributeValuesPerObjectAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attribute/get"
	var error = "orion/server/misc/error/attribute/get"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "GetAttributeValuesPerObjectAction",
		Description:    "Get attribute values based on objectId (mandatory) and attributeId (optional)",
		RequestPath:    "orion/server/misc/request/attribute/get",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.baseAction.ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   true,
	}

	return info
}

func (action *GetAttributeValuesPerObjectAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetAttributeValuesPerObjectAction) createGetAttributeDefinitionsReply(attributes []structs2.Attribute) (structs.GetAttributeValuesReply, *micro.Exception) {
	var reply = structs.GetAttributeValuesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(attributes) > 0 {
		reply.Header.Success = true
		reply.Attributes = attributes
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No attributes were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetAttributeValuesPerObjectAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetAttributeValuesRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	reply, myErr := action.getAttributeValues(receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetAttributeValuesPerObjectAction) fillAttributeValues(rows *sql.Rows) ([]structs2.Attribute, *micro.Exception) {
	var attributes []structs2.Attribute

	for rows.Next() {
		var attribute structs2.Attribute
		/*
			attr_id, attr_value, a.object_type, object_id, b.datatype, b.name, b.description, " +
			"(extract(epoch from b.created_date)::bigint)*1000 AS created_date, " +
			"b.active, b.pretty_id, a.id, b.overwriteable "
		*/
		err := rows.Scan(&attribute.AttributeId, &attribute.Value, &attribute.ObjectType, &attribute.ObjectId,
			&attribute.DataType, &attribute.Info.Name, &attribute.Info.Description,
			&attribute.Info.CreatedDate, &attribute.Info.Active, &attribute.Info.Alias, &attribute.Info.Id,
			&attribute.Overwriteable)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		attributes = append(attributes, attribute)
	}

	return attributes, nil
}

func (action GetAttributeValuesPerObjectAction) getAttributeValues(request structs.GetAttributeValuesRequest) (structs.GetAttributeValuesReply, *micro.Exception) {
	attributes, myErr := action.getAttributeValuesFromDb(request)

	if myErr != nil {
		return structs.GetAttributeValuesReply{}, myErr
	}

	return action.createGetAttributeDefinitionsReply(attributes)
}

func (action GetAttributeValuesPerObjectAction) getAttributesPerObjectId(objectId *uint64, sql string) ([]structs2.Attribute, *micro.Exception) {
	rows, err := action.GetBaseAction().Environment.Database.Query(sql, objectId)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()

	attributes, myErr := action.fillAttributeValues(rows)
	if myErr != nil {
		return attributes, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return attributes, nil
}

func (action GetAttributeValuesPerObjectAction) getAttributesPerObjectIdAndAttributeId(objectId, attributeId *uint64, sql string) ([]structs2.Attribute, *micro.Exception) {
	rows, err := action.GetBaseAction().Environment.Database.Query(sql, objectId, attributeId)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()

	attributes, myErr := action.fillAttributeValues(rows)
	if myErr != nil {
		return attributes, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return attributes, nil
}

func (action GetAttributeValuesPerObjectAction) getAttributeValuesFromDb(request structs.GetAttributeValuesRequest) ([]structs2.Attribute, *micro.Exception) {
	var sql = SqlGetAttributeValues

	if request.AttributeId != nil {
		sql += " AND a.attr_id = $2 "
	}
	logger := logging.GetLogger("GetAttributeValuesPerObjectAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetAttributeValuesPerObjectAction query")

	if request.AttributeId != nil {
		return action.getAttributesPerObjectIdAndAttributeId(request.ObjectId, request.AttributeId, sql)
	}

	return action.getAttributesPerObjectId(request.ObjectId, sql)
}
