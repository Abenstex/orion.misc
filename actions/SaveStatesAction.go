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
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"orion.misc/structs"
	"time"
)

type SaveStatesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.State
	startedTime  int64
}

func (action SaveStatesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveStatesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Errorf("error unmarshalling the request: %v\n", string(request))
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)
	if err != nil {
		return micro.NewException(structs2.RequestHeaderInvalid, err)
	}

	return nil
}

func (action SaveStatesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveStatesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveStatesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveStatesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveStatesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveStatesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveStatesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveStatesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveStatesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}

	event := structs.SavedStatesEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		States:     action.savedObjects,
		ObjectType: "STATE",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveStatesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveStatesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/state/save"
	var error = "orion/server/misc/error/state/save"
	var event = "orion/server/misc/event/state/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "SaveStatesAction",
		Description:     "Saves states to the database",
		RequestTopic:    "orion/server/misc/request/state/save",
		ReplyTopic:      reply,
		ErrorReplyTopic: error,
		Version:         1,
		ClientId:        action.GetBaseAction().ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		EventTopic:      event,
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		IsScriptable:    false,
	}

	return info
}

func (action *SaveStatesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveStatesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = utils2.GetCurrentTimeStamp()

	saveRequest := structs.SaveStatesRequest{}

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	orionErr := action.saveObjects(ctx, saveRequest.UpdatedStates, saveRequest.Header.Comment, saveRequest.Header.User)
	if orionErr != nil {
		logging.GetLogger(action.ProvideInformation().Name,
			action.GetBaseAction().Environment,
			true).WithError(err).Error("Data could not be saved")
		return structs2.NewErrorReplyHeaderWithOrionErr(orionErr,
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	reply := structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveStatesAction) archiveAndReplaceObject(ctx context.Context, object structs2.State) error {
	var objectToArchive structs2.State
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "states", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	objectToArchive.ID = nil
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "states", objectToArchive)

	return err
}

func (action *SaveStatesAction) saveObjects(ctx context.Context, objects []structs2.State, comment, user string) *structs2.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs2.State)
		for _, object := range objects {
			object.Info.UserComment = &comment
			object.Info.User = &user
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = utils2.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "states", object)
				if err != nil {
					return nil, err
				}
			} else {

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
		return structs2.NewOrionError(structs2.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
