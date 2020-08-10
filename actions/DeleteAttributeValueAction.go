package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	"laniakea/mqtt"
	laniakea "laniakea/utils"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	structs2 "orion.commons/structs"
	"orion.commons/utils"
	"orion.misc/structs"
	"time"
)

type DeleteAttributeValueAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteAttributeValueRequest
}

func (action *DeleteAttributeValueAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &action.deleteRequest.Header, action, true)
	if err != nil {
		return micro.NewException(structs2.RequestHeaderInvalid, err)
	}

	return nil
}

func (action *DeleteAttributeValueAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteAttributeValueAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteAttributeValueAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteAttributeValueAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteAttributeValueAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteAttributeValueAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteAttributeValueAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteAttributeValueRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteAttributeValueAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(delRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs.AttributeValueDeletedEvent{
		Header:      *micro.NewEventHeaderForAction(action.ProvideInformation(), delRequest.Header.SenderId, ""),
		AttributeId: action.deleteRequest.AttributeId,
		ObjectId:    action.deleteRequest.ObjectId,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteAttributeValueAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteAttributeValueAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attribute/delete"
	var error = "orion/server/misc/error/attribute/delete"
	var event = "orion/server/misc/event/attribute/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteAttributeValueRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs.AttributeValueDeletedEvent{})
	info := micro.ActionInformation{
		Name:           "DeleteAttributeValueAction",
		Description:    "Deletes an attribute/object/value combination from the database",
		RequestPath:    "orion/server/misc/request/attribute/delete",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		EventSample:    dataStructures.JsonNullString{NullString: sql.NullString{String: eventSample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *DeleteAttributeValueAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteAttributeValueAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	err = action.deleteAttributeValue(action.deleteRequest.AttributeId, action.deleteRequest.ObjectId)
	if err != nil {
		logging.GetLogger("DeleteAttributeValueAction", action.baseAction.Environment, true).
			WithError(err).
			Error("attribute value with attribute id %v and object id %v could not be deleted from the database", action.deleteRequest.AttributeId, action.deleteRequest.ObjectId)

		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteAttributeValueAction) deleteAttributeValue(attributeId, objectId uint64) error {
	var sql = "DELETE FROM ref_attributes_objects WHERE attr_id=$1 AND object_id=$2"
	logger := logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true)
	logger.WithFields(logrus.Fields{
		"DELETE statement": sql,
	}).Debug("Deleting attribute value: ")

	return laniakea.ExecuteQueryInTransaction(action.baseAction.Environment, sql, attributeId, objectId)
}
