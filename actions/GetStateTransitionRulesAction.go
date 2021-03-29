package actions

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/micro"
	utils2 "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"orion.misc/structs"
	"time"
)

type GetStateTransitionRulesAction struct {
	baseAction      micro.BaseAction
	MetricsStore    *utils.MetricsStore
	receivedRequest structs.GetStateTransitionRulesRequest
}

func (action *GetStateTransitionRulesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetStateTransitionRulesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	action.receivedRequest = dummy
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetStateTransitionRulesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetStateTransitionRulesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetStateTransitionRulesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetStateTransitionRulesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetStateTransitionRulesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetStateTransitionRulesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetStateTransitionRulesAction) SendEvents(request micro.IRequest) {

}

func (action GetStateTransitionRulesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/statetransitionrule/get"
	var error = "orion/server/misc/error/statetransitionrule/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetStateTransitionRulesRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetStateTransitionRulesReply{})
	info := micro.ActionInformation{
		Name:            "GetStateTransitionRulesAction",
		Description:     "Gets all state transition rules from the database",
		RequestTopic:    "orion/server/misc/request/statetransitionrule/get",
		ReplyTopic:      reply,
		ErrorReplyTopic: error,
		Version:         1,
		ClientId:        action.baseAction.ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		IsScriptable:    false,
	}

	return info
}

func (action *GetStateTransitionRulesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetStateTransitionRulesAction) createGetStateTransitionRulesReply(objects []structs.StateTransitionRule) (structs.GetStateTransitionRulesReply, *structs2.OrionError) {
	var reply = structs.GetStateTransitionRulesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(objects) > 0 {
		reply.Header.Success = true
		reply.StateTransitionRules = objects
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No objects were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetStateTransitionRulesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &action.receivedRequest
	}

	reply, myErr := action.getStateTransitionRules(ctx, action.receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &action.receivedRequest
	}

	return reply, &action.receivedRequest
}

func (action GetStateTransitionRulesAction) getStateTransitionRules(ctx context.Context, request structs.GetStateTransitionRulesRequest) (structs.GetStateTransitionRulesReply, *structs2.OrionError) {
	objects, myErr := action.getStateTransitionRulesFromDb(ctx, request)

	if myErr != nil {
		return structs.GetStateTransitionRulesReply{}, myErr
	}

	return action.createGetStateTransitionRulesReply(objects)
}

func (action GetStateTransitionRulesAction) getStateTransitionRulesFromDb(ctx context.Context, request structs.GetStateTransitionRulesRequest) ([]structs.StateTransitionRule, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("attribute_definitions").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs.StateTransitionRule
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
