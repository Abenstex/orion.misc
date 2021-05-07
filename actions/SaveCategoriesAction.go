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
	laniakea "github.com/abenstex/laniakea/utils"
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

type SaveCategoriesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.Category
	saveRequest  structs2.SaveCategoriesRequest
	startedTime  int64
}

func (action *SaveCategoriesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveCategoriesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)

	action.saveRequest = dummy

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action SaveCategoriesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveCategoriesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveCategoriesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveCategoriesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveCategoriesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveCategoriesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveCategoriesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveCategoriesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveCategoriesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}

	event := structs2.CategorySavedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Categories: action.savedObjects,
		ObjectType: "CATEGORY",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveCategoriesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveCategoriesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/save"
	var error = "orion/server/misc/error/category/save"
	var event = "orion/server/misc/event/category/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveCategoriesRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.CategorySavedEvent{})
	info := micro.ActionInformation{
		Name:            "SaveCategoriesAction",
		Description:     "Saves categories and all necessary references to the database",
		RequestTopic:    "orion/server/misc/request/category/save",
		ReplyTopic:      reply,
		ErrorReplyTopic: error,
		Version:         1,
		ClientId:        action.GetBaseAction().ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		EventTopic:      event,
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		EventSample:     &eventSample,
		IsScriptable:    false,
	}

	return info
}

func (action *SaveCategoriesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveCategoriesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = laniakea.GetCurrentTimeStamp()

	dummy, _ := json.Marshal(action.saveRequest)
	fmt.Printf("Request: %v\n", string(dummy))

	exception := action.saveObjects(ctx, action.saveRequest.UpdatedCategories, action.saveRequest.Header.Comment, action.saveRequest.Header.User)
	if exception != nil {
		logging.GetLogger("SaveCategoriesAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(exception,
			action.ProvideInformation().ErrorReplyTopic), &action.saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &action.saveRequest
}

func (action *SaveCategoriesAction) archiveAndReplaceObject(ctx context.Context, object structs2.Category) error {
	var objectToArchive structs2.Category
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "categories", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	objectToArchive.ID = nil
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "categories", objectToArchive)

	return err
}

func (action *SaveCategoriesAction) saveObjects(ctx context.Context, objects []structs2.Category, comment, user string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs2.Category)
		for _, object := range objects {
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = laniakea.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "categories", object)
				if err != nil {
					return nil, err
				}
			} else {
				object.Info.UserComment = &comment
				object.Info.User = &user
				object.Info.ChangeDate = &action.startedTime

				err := action.archiveAndReplaceObject(sessCtx, object)
				if err != nil {
					return nil, err
				}
			}
			action.savedObjects = append(action.savedObjects, object)
		}

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs.NewOrionError(structs.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
