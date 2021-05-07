package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mongodb"
	"github.com/abenstex/laniakea/mqtt"
	utils2 "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	"github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type DeleteAttributeDefinitionAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
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
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs.DeletedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), delRequest.Header.SenderId, ""),
		ObjectId:   delRequest.ObjectId,
		ObjectType: "ATTRIBUTE",
		ObjectName: delRequest.ObjectName,
	}

	jsonString, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteAttributeDefinitionAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, jsonString, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteAttributeDefinitionAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/delete"
	var errorSubject = "orion/server/misc/error/attributedefinition/delete"
	var event = "orion/server/misc/event/attributedefinition/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "DeleteAttributeDefinitionAction",
		Description:     "Delete an attribute definition from the database",
		RequestTopic:    "orion/server/misc/request/attributedefinition/delete",
		ReplyTopic:      reply,
		ErrorReplyTopic: errorSubject,
		Version:         1,
		ClientId:        action.GetBaseAction().ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		EventTopic:      event,
		IsScriptable:    false,
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

	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &action.deleteRequest
	}

	//err = utils.DeleteObjectById(env, "attributes", action.deleteRequest.ObjectId, "DeleteAttributeDefinitionAction")
	orionErr := action.deleteObject(ctx, action.deleteRequest.ObjectId)
	if orionErr != nil {
		logging.GetLogger("DeleteAttributeDefinitionAction", action.baseAction.Environment, true).
			WithError(orionErr.Error).
			Error("object with id %v could not be deleted from the database", action.deleteRequest.ObjectId)

		return structs.NewErrorReplyHeaderWithOrionErr(orionErr,
			action.ProvideInformation().ErrorReplyTopic), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteAttributeDefinitionAction) deleteObject(ctx context.Context, id string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "id", id)
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		callbackId := fmt.Sprintf("%v", newCtx.Value("id"))
		result, err := mongodb.DeleteAndFindOneById(sessCtx, action.baseAction.Environment.MongoDbConnection, "attribute_definitions", callbackId)
		if err != nil {
			return nil, err
		}
		var objectToArchive structs.AttributeDefinition
		err = result.Decode(&objectToArchive)
		if err != nil {
			return nil, err
		}
		time := utils2.GetCurrentTimeStamp()
		objectToArchive.Info.DeletionDate = &time
		objectToArchive.ID = nil
		_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "attribute_definitions", objectToArchive)

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs.NewOrionError(structs.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
