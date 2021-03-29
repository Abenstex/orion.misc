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
	structs2 "orion.misc/structs"
	"time"
)

type DeleteStateAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
	objectName    string
}

func (action *DeleteStateAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &action.deleteRequest.Header, action, true)
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}
	action.objectName = action.deleteRequest.ObjectName

	return nil
}

func (action *DeleteStateAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteStateAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteStateAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteStateAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteStateAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteStateAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteStateAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteStateAction",
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
		ObjectType: "STATE",
		ObjectName: action.objectName,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteStateAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteStateAction) ProvideInformation() micro.ActionInformation {
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs.DeletedEvent{})
	info := micro.ActionInformation{
		Name:            "DeleteStateAction",
		Description:     "Delete a state from the database",
		RequestTopic:    "orion/server/misc/request/state/delete",
		ReplyTopic:      "orion/server/misc/reply/state/delete",
		ErrorReplyTopic: "orion/server/misc/error/state/delete",
		Version:         1,
		ClientId:        action.GetBaseAction().ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		EventTopic:      "orion/server/misc/event/state/delete",
		EventSample:     &eventSample,
		IsScriptable:    false,
	}

	return info
}

func (action *DeleteStateAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteStateAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyTopic), &action.deleteRequest
	}

	orionErr := action.deleteObject(ctx, action.deleteRequest.ObjectId)
	if orionErr != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(orionErr,
			action.ProvideInformation().ErrorReplyTopic), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteStateAction) deleteObject(ctx context.Context, id string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "id", id)
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		callbackId := fmt.Sprintf("%v", newCtx.Value("id"))
		result, err := mongodb.DeleteAndFindOneById(sessCtx, action.baseAction.Environment.MongoDbConnection, "states", callbackId)
		if err != nil {
			return nil, err
		}
		var objectToArchive structs2.Category
		err = result.Decode(&objectToArchive)
		if err != nil {
			return nil, err
		}
		time := utils2.GetCurrentTimeStamp()
		objectToArchive.Info.DeletionDate = &time

		_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "states", objectToArchive)

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs.NewOrionError(structs.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
