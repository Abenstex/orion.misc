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

type SaveParametersAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.Parameter
	startedTime  int64
}

func (action SaveParametersAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveParametersRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action SaveParametersAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveParametersAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveParametersAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveParametersAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveParametersAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveParametersAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveParametersAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveParametersRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveParametersAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}

	event := structs2.ParameterSavedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Parameters: action.savedObjects,
		ObjectType: "PARAMETER",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveParametersAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveParametersAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/parameter/save"
	var error = "orion/server/misc/error/parameter/save"
	var event = "orion/server/misc/event/parameter/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveParametersRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.ParameterSavedEvent{})
	info := micro.ActionInformation{
		Name:            "SaveParametersAction",
		Description:     "Saves PARAMETER and all necessary references to the database",
		RequestTopic:    "orion/server/misc/request/parameter/save",
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

func (action *SaveParametersAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveParametersAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = laniakea.GetCurrentTimeStamp()

	saveRequest := structs2.SaveParametersRequest{}

	err := json.Unmarshal(request, &saveRequest)
	//fmt.Printf("Saverequest: %v\n", string(request))
	if err != nil {
		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	exception := action.saveObjects(ctx, saveRequest.Parameters, saveRequest.Header.Comment, saveRequest.Header.User)
	if exception != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SaveParametersAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(exception,
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveParametersAction) archiveAndReplaceObject(ctx context.Context, object structs2.Parameter) error {
	var objectToArchive structs2.Parameter
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "parameters", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "parameters", objectToArchive)

	return err
}

func (action *SaveParametersAction) saveObjects(ctx context.Context, objects []structs2.Parameter, comment, user string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs2.Parameter)
		for _, object := range objects {
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = laniakea.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "parameters", object)
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
