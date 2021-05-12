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

type GetAttributeDefinitionsAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetAttributeDefinitionsAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetAttributeDefinitionsRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	if err != nil {
		return micro.NewException(structs2.RequestHeaderInvalid, err)
	}

	return nil
}

func (action GetAttributeDefinitionsAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetAttributeDefinitionsAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetAttributeDefinitionsAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetAttributeDefinitionsAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetAttributeDefinitionsAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetAttributeDefinitionsAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetAttributeDefinitionsAction) SendEvents(request micro.IRequest) {

}

func (action GetAttributeDefinitionsAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/get"
	var errorTopic = "orion/server/misc/error/attributedefinition/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetAttributeDefinitionsRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetAttributeDefinitionsReply{})
	info := micro.ActionInformation{
		Name:            "GetAttributeDefinitionsAction",
		Description:     "Get attribute definitions based on conditions or all if no conditions were sent in the request",
		RequestTopic:    "orion/server/misc/request/attributedefinition/get",
		ReplyTopic:      reply,
		ErrorReplyTopic: errorTopic,
		Version:         1,
		ClientId:        action.baseAction.ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		IsScriptable:    false,
	}

	return info
}

func (action *GetAttributeDefinitionsAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetAttributeDefinitionsAction) createGetAttributeDefinitionsReply(definitions []structs2.AttributeDefinition) (structs.GetAttributeDefinitionsReply, *structs2.OrionError) {
	var reply = structs.GetAttributeDefinitionsReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(definitions) > 0 {
		reply.Header.Success = true
		reply.AttributeDefinitions = definitions
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No definitions were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetAttributeDefinitionsAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetAttributeDefinitionsRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	reply, myErr := action.getAttributeDefinitions(ctx, receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetAttributeDefinitionsAction) getAttributeDefinitions(ctx context.Context, request structs.GetAttributeDefinitionsRequest) (structs.GetAttributeDefinitionsReply, *structs2.OrionError) {
	attributes, myErr := action.getAttributeDefinitionsFromDb(ctx, request)

	if myErr != nil {
		return structs.GetAttributeDefinitionsReply{}, myErr
	}

	return action.createGetAttributeDefinitionsReply(attributes)
}

func (action GetAttributeDefinitionsAction) getAttributeDefinitionsFromDb(ctx context.Context, request structs.GetAttributeDefinitionsRequest) ([]structs2.AttributeDefinition, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("attribute_definitions").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs2.AttributeDefinition
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
