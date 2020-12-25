package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mqtt"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	"github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type DeleteObjectTypeCustomizationAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
}

func (action *DeleteObjectTypeCustomizationAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &action.deleteRequest.Header, action, true)
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}

	return nil
}

func (action *DeleteObjectTypeCustomizationAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteObjectTypeCustomizationAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteObjectTypeCustomizationAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteObjectTypeCustomizationAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteObjectTypeCustomizationAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteObjectTypeCustomizationAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteObjectTypeCustomizationAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteObjectTypeCustomizationAction",
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
		ObjectType: "OBJECT_TYPE_CUSTOMIZATION",
		ObjectName: "",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteObjectTypeCustomizationAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteObjectTypeCustomizationAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/objectcustomization/delete"
	var error = "orion/server/misc/error/objectcustomization/delete"
	var event = "orion/server/misc/event/objectcustomization/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "DeleteObjectTypeCustomizationAction",
		Description:    "Delete an object type customization definition from the database",
		RequestPath:    "orion/server/misc/request/objectcustomization/delete",
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

func (action *DeleteObjectTypeCustomizationAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteObjectTypeCustomizationAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	env := action.GetBaseAction().Environment
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	err = utils.DeleteObjectById(env, "object_type_customizations", action.deleteRequest.ObjectId, "DeleteObjectTypeCustomizationAction")
	if err != nil {
		logging.GetLogger("DeleteObjectTypeCustomizationAction", action.baseAction.Environment, true).
			WithError(err).
			Error("object with id %v could not be deleted from the database", action.deleteRequest.ObjectId)

		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}
