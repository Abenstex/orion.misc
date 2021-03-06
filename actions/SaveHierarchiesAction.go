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
	http2 "github.com/abenstex/orion.commons/http"
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"orion.misc/structs"
	"time"
)

type SaveHierarchiesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs.Hierarchy
	startedTime  int64
}

func (action SaveHierarchiesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveHierarchiesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Errorf("error unmarshalling the request: %v\n", string(request))
		return micro.NewException(structs2.UnmarshalError, err)
	}
	if err != nil {
		return micro.NewException(structs2.RequestHeaderInvalid, err)
	}

	return nil
}

func (action SaveHierarchiesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveHierarchiesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveHierarchiesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveHierarchiesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveHierarchiesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveHierarchiesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveHierarchiesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveHierarchiesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveHierarchiesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}

	event := structs.SavedHierarchiesEvent{
		Header:      *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Hierarchies: action.savedObjects,
		ObjectType:  "HIERARCHY",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveHierarchiesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveHierarchiesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/hierarchy/save"
	var error = "orion/server/misc/error/hierarchy/save"
	var event = "orion/server/misc/event/hierarchy/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs.SavedHierarchiesEvent{})
	info := micro.ActionInformation{
		Name:            "SaveHierarchiesAction",
		Description:     "Saves hierarchies to the database",
		RequestTopic:    "orion/server/misc/request/hierarchy/save",
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

func (action *SaveHierarchiesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveHierarchiesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs.SaveHierarchiesRequest{}
	action.startedTime = utils2.GetCurrentTimeStamp()

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	exception := action.saveObjects(ctx, saveRequest.UpdatedHierarchies, saveRequest.Header.Comment, saveRequest.Header.User)
	if exception != nil {
		logging.GetLogger(action.ProvideInformation().Name,
			action.GetBaseAction().Environment,
			true).WithField("exception", exception).Error("Data could not be saved")
		return structs2.NewErrorReplyHeaderWithOrionErr(exception,
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	reply := structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveHierarchiesAction) archiveAndReplaceObject(ctx context.Context, object structs.Hierarchy) error {
	var objectToArchive structs.Hierarchy
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "hierarchies", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	objectToArchive.ID = nil
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "hierarchies", objectToArchive)

	return err
}

func (action *SaveHierarchiesAction) saveObjects(ctx context.Context, objects []structs.Hierarchy, comment, user string) *structs2.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs.Hierarchy)
		for _, object := range objects {
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = utils2.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "hierarchies", object)
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
		return structs2.NewOrionError(structs2.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
