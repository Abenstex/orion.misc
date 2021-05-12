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

type SaveObjectTypeCustomizationAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.ObjectTypeCustomization
	saveRequest  structs2.SaveObjectTypeCustomizationsRequest
	startedTime  int64
}

func (action *SaveObjectTypeCustomizationAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveObjectTypeCustomizationsRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)
	action.saveRequest = dummy
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}

	return nil
}

func (action SaveObjectTypeCustomizationAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveObjectTypeCustomizationAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveObjectTypeCustomizationAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveObjectTypeCustomizationAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveObjectTypeCustomizationAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveObjectTypeCustomizationAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveObjectTypeCustomizationAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveObjectTypeCustomizationsRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveObjectTypeCustomizationAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]string, 0, len(action.savedObjects))
	for _, parameter := range action.savedObjects {
		ids = append(ids, parameter.ID.Hex())
	}
	event := structs2.ObjectTypeCustomizationsSavedEvent{
		Header:                   *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		ObjectTypeCustomizations: ids,
		ObjectType:               "OBJECT_TYPE_CUSTOMIZATION",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveObjectTypeCustomizationAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveObjectTypeCustomizationAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/objectcustomization/save"
	var error = "orion/server/misc/error/objectcustomization/save"
	var event = "orion/server/misc/event/objectcustomization/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveObjectTypeCustomizationsRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.ObjectTypeCustomizationsSavedEvent{})
	info := micro.ActionInformation{
		Name:            "SaveObjectTypeCustomizationAction",
		Description:     "Saves Object Type Customizations and all necessary references to the database",
		RequestTopic:    "orion/server/misc/request/objectcustomization/save",
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

func (action *SaveObjectTypeCustomizationAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveObjectTypeCustomizationAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = laniakea.GetCurrentTimeStamp()

	exception := action.saveObjects(ctx, action.saveRequest.ObjectTypeCustomizations, action.saveRequest.Header.Comment, action.saveRequest.Header.User)
	if exception != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SaveObjectTypeCustomizationAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(exception,
			action.ProvideInformation().ErrorReplyTopic), &action.saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &action.saveRequest
}

func (action *SaveObjectTypeCustomizationAction) archiveAndReplaceObject(ctx context.Context, object structs2.ObjectTypeCustomization) error {
	var objectToArchive structs2.ObjectTypeCustomization
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "object_type_customization", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.ChangeDate = &action.startedTime
	objectToArchive.ID = nil
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "object_type_customization", objectToArchive)

	return err
}

func (action *SaveObjectTypeCustomizationAction) saveObjects(ctx context.Context, objects []structs2.ObjectTypeCustomization, comment, user string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs2.ObjectTypeCustomization)
		for _, object := range objects {
			if object.CreatedDate == 0 {
				object.CreatedDate = laniakea.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "object_type_customization", object)
				if err != nil {
					return nil, err
				}
			} else {
				object.UserComment = &comment
				object.User = &user
				object.ChangeDate = &action.startedTime

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
