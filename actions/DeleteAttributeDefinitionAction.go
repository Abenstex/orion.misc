package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/spf13/viper"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	"laniakea/mqtt"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	"orion.commons/structs"
	"orion.commons/utils"
	"time"
)

type DeleteAttributeDefinitionAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
	objectName    string
}

func (action *DeleteAttributeDefinitionAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &action.deleteRequest.Header, action, true)
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}
	err = action.getNameBeforeDelete(action.deleteRequest.ObjectId)
	if err != nil {
		return micro.NewException(structs.DatabaseError, err)
	}

	return nil
}

func (action *DeleteAttributeDefinitionAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteAttributeDefinitionAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteAttributeDefinitionAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteAttributeDefinitionAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteAttributeDefinitionAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteAttributeDefinitionAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteAttributeDefinitionAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteAttributeDefinitionAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(delRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs.DeletedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), delRequest.Header.SenderId, ""),
		ObjectId:   delRequest.ObjectId,
		ObjectType: "ATTRIBUTE",
		ObjectName: action.objectName,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteAttributeDefinitionAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteAttributeDefinitionAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/delete"
	var error = "orion/server/misc/error/attributedefinition/delete"
	var event = "orion/server/misc/event/attributedefinition/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "DeleteAttributeDefinitionAction",
		Description:    "Delete an attribute definition from the database",
		RequestPath:    "orion/server/misc/request/attributedefinition/delete",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *DeleteAttributeDefinitionAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteAttributeDefinitionAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	env := action.GetBaseAction().Environment
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	err = utils.DeleteObjectById(env, "attributes", action.deleteRequest.ObjectId, "DeleteAttributeDefinitionAction")
	if err != nil {
		logging.GetLogger("DeleteAttributeDefinitionAction", action.baseAction.Environment, true).
			WithError(err).
			Error("object with id %v could not be deleted from the database", action.deleteRequest.ObjectId)

		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteAttributeDefinitionAction) getNameBeforeDelete(objectId int64) error {
	query := "SELECT name FROM attributes WHERE id=$1"
	var name string
	row := action.GetBaseAction().Environment.Database.QueryRow(query, objectId)
	err := row.Scan(&name)
	if err != nil {
		logging.GetLogger("DeleteAttributeDefinitionAction", action.baseAction.Environment, true).
			WithError(err).
			Error("Could not get name for object before delete (ID: %v", objectId)
	}

	action.objectName = name

	return err
}
