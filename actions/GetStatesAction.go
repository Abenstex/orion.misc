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

type GetStatesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetStatesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetStatesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetStatesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetStatesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetStatesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetStatesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetStatesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetStatesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetStatesAction) SendEvents(request micro.IRequest) {

}

const SqlGetAllStates = "SELECT id, name, description, active, (extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, " +
	" b.action_by, referenced_type, object_available, substate, default_state " +
	" FROM states a left outer join cache b on a.id=b.object_id "

func (action GetStatesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/state/get"
	var error = "orion/server/misc/error/state/get"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "GetStatesAction",
		Description:     "Get states based on conditions or all if no conditions were sent in the request",
		RequestTopic:    "orion/server/misc/request/state/get",
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

func (action *GetStatesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetStatesAction) createGetStatesReply(states []structs2.State) (structs.GetStatesReply, *structs2.OrionError) {
	var reply = structs.GetStatesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(states) > 0 {
		reply.Header.Success = true
		reply.States = states
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No states were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetStatesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetStatesRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	reply, myErr := action.getStates(ctx, receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetStatesAction) getStates(ctx context.Context, request structs.GetStatesRequest) (structs.GetStatesReply, *structs2.OrionError) {
	states, myErr := action.getStatesFromDb(ctx, request)

	if myErr != nil {
		return structs.GetStatesReply{}, myErr
	}

	return action.createGetStatesReply(states)
}

func (action GetStatesAction) getStatesFromDb(ctx context.Context, request structs.GetStatesRequest) ([]structs2.State, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("states").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs2.State
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
