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

type GetHierarchiesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetHierarchiesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetStatesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetHierarchiesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetHierarchiesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetHierarchiesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetHierarchiesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetHierarchiesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetHierarchiesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetHierarchiesAction) SendEvents(request micro.IRequest) {

}

func (action GetHierarchiesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/hierarchy/get"
	var error = "orion/server/misc/error/hierarchy/get"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "GetHierarchiesAction",
		Description:     "Get hierarchies based on conditions or all if no conditions were sent in the request",
		RequestTopic:    "orion/server/misc/request/hierarchy/get",
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

func (action *GetHierarchiesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetHierarchiesAction) createGetHierarchiesReply(hierarchies []structs.Hierarchy) (structs.GetHierarchiesReply, *structs2.OrionError) {
	var reply = structs.GetHierarchiesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(hierarchies) > 0 {
		reply.Header.Success = true
		reply.Hierarchies = hierarchies
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No hierarchies were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetHierarchiesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetHierarchiesRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	reply, myErr := action.getHierarchies(ctx, receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetHierarchiesAction) getHierarchies(ctx context.Context, request structs.GetHierarchiesRequest) (structs.GetHierarchiesReply, *structs2.OrionError) {
	hierarchies, myErr := action.getHierarchiesFromDb(ctx, request)

	if myErr != nil {
		return structs.GetHierarchiesReply{}, myErr
	}

	return action.createGetHierarchiesReply(hierarchies)
}

func (action GetHierarchiesAction) getHierarchiesFromDb(ctx context.Context, request structs.GetHierarchiesRequest) ([]structs.Hierarchy, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("hierarchies").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs.Hierarchy
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
