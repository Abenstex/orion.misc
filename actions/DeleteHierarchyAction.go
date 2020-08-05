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

type DeleteHierarchyAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
	objectName    string
}

func (action *DeleteHierarchyAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
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

func (action *DeleteHierarchyAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteHierarchyAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteHierarchyAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteHierarchyAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteHierarchyAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteHierarchyAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteHierarchyAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteHierarchyAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(delRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs.DeletedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), request.GetHeader().SenderId, ""),
		ObjectId:   delRequest.ObjectId,
		ObjectType: "HIERARCHY",
		ObjectName: action.objectName,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteHierarchyAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteHierarchyAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/hierarchy/delete"
	var error = "orion/server/misc/error/hierarchy/delete"
	var event = "orion/server/misc/event/hierarchy/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "DeleteHierarchyAction",
		Description:    "Delete a hierarchy from the database",
		RequestPath:    "orion/server/misc/request/hierarchy/delete",
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

func (action *DeleteHierarchyAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteHierarchyAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	env := action.GetBaseAction().Environment
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	err = utils.DeleteObjectById(env, "hierarchies", action.deleteRequest.ObjectId, "DeleteHierarchyAction")
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteHierarchyAction) getNameBeforeDelete(objectId int64) error {
	query := "SELECT name FROM hierarchies WHERE id=$1"
	var name string
	row := action.GetBaseAction().Environment.Database.QueryRow(query, objectId)
	err := row.Scan(&name)

	action.objectName = name

	return err
}
