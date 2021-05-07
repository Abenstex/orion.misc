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

type SaveStateTransitionRulesAction struct {
	baseAction                micro.BaseAction
	MetricsStore              *utils.MetricsStore
	savedStateTransitionRules []structs.StateTransitionRule
	startedTime               int64
}

func (action SaveStateTransitionRulesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveStateTransitionRulesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action SaveStateTransitionRulesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveStateTransitionRulesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveStateTransitionRulesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveStateTransitionRulesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveStateTransitionRulesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveStateTransitionRulesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveStateTransitionRulesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveStateTransitionRulesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveStateTransitionRulesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}

	event := structs.SavedStateTransitionRulesEvent{
		Header:               *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		StateTransitionRules: action.savedStateTransitionRules,
		ObjectType:           "STATE_TRANSITION_RULE",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveStateTransitionRulesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveStateTransitionRulesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/statetransitionrule/save"
	var error = "orion/server/misc/error/statetransitionrule/save"
	var event = "orion/server/misc/event/statetransitionrule/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "SaveStateTransitionRulesAction",
		Description:     "Saves statetransitionrules to the database",
		RequestTopic:    "orion/server/misc/request/statetransitionrule/save",
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

func (action *SaveStateTransitionRulesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveStateTransitionRulesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = utils2.GetCurrentTimeStamp()

	saveRequest := structs.SaveStateTransitionRulesRequest{}

	err := json.Unmarshal(request, &saveRequest)
	//fmt.Printf("Saverequest: %v\n", string(request))
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	orionErr := action.saveObjects(ctx, saveRequest.UpdatedStateTransitionRules, saveRequest.Header.Comment, saveRequest.Header.User)
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

func (action *SaveStateTransitionRulesAction) archiveAndReplaceObject(ctx context.Context, object structs.StateTransitionRule) error {
	var objectToArchive structs.StateTransitionRule
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "state_transition_rules", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	objectToArchive.ID = nil
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "state_transition_rules", objectToArchive)

	return err
}

func (action *SaveStateTransitionRulesAction) saveObjects(ctx context.Context, objects []structs.StateTransitionRule, comment, user string) *structs2.OrionError {
	newCtx := context.WithValue(ctx, "objects", objects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs.StateTransitionRule)
		for _, object := range objects {
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = utils2.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "state_transition_rules", object)
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
			action.savedStateTransitionRules = append(action.savedStateTransitionRules, object)
		}

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs2.NewOrionError(structs2.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}
