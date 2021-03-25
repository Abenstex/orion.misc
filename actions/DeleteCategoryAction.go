package actions

import (
	"context"
	"database/sql"
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

type DeleteCategoryAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
	objectName    string
}

func (action *DeleteCategoryAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
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

func (action *DeleteCategoryAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *DeleteCategoryAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *DeleteCategoryAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action *DeleteCategoryAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteCategoryAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteCategoryAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteCategoryAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteCategoryAction",
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
		ObjectType: "CATEGORY",
		ObjectName: action.objectName,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteCategoryAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteCategoryAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/delete"
	var error = "orion/server/misc/error/category/delete"
	var event = "orion/server/misc/event/category/delete"
	var requestSample = dataStructures.StructToJsonString(structs.DeleteRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs.DeletedEvent{})
	info := micro.ActionInformation{
		Name:           "DeleteCategoryAction",
		Description:    "Delete a category from the database",
		RequestPath:    "orion/server/misc/request/category/delete",
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

func (action *DeleteCategoryAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteCategoryAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	orionErr := action.deleteObject(ctx, action.deleteRequest.ObjectId)
	if orionErr != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(orionErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteCategoryAction) deleteObject(ctx context.Context, id string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "id", id)
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		callbackId := fmt.Sprintf("%v", newCtx.Value("id"))
		result, err := mongodb.DeleteAndFindOneById(sessCtx, action.baseAction.Environment.MongoDbConnection, "categories", callbackId)
		if err != nil {
			return nil, err
		}
		var objectToArchive structs2.Category
		err = result.Decode(&objectToArchive)
		if err != nil {
			return nil, err
		}
		objectToArchive.Info.DeletionDate = dataStructures.JsonNullInt64{NullInt64: sql.NullInt64{
			Int64: utils2.GetCurrentTimeStamp(),
			Valid: true,
		}}
		_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "categories", objectToArchive)

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs.NewOrionError(structs.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
