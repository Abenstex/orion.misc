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

type GetObjectTypeCustomizationsAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	request      structs.GetObjectTypeCustomizationsRequest
}

func (action GetObjectTypeCustomizationsAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetObjectTypeCustomizationsRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	action.request = dummy
	if err != nil {
		return micro.NewException(structs2.RequestHeaderInvalid, err)
	}

	return nil
}

func (action GetObjectTypeCustomizationsAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetObjectTypeCustomizationsAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetObjectTypeCustomizationsAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetObjectTypeCustomizationsAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetObjectTypeCustomizationsAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetObjectTypeCustomizationsAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetObjectTypeCustomizationsAction) SendEvents(request micro.IRequest) {

}

func (action GetObjectTypeCustomizationsAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/objectcustomization/get"
	var error = "orion/server/misc/error/objectcustomization/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetCategoriesRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetCategoriesReply{})
	info := micro.ActionInformation{
		Name:            "GetObjectTypeCustomizationsAction",
		Description:     "Get object type customizations based on conditions or all if no conditions were sent in the request",
		RequestTopic:    "orion/server/misc/request/objectcustomization/get",
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

func (action *GetObjectTypeCustomizationsAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetObjectTypeCustomizationsAction) createGetObjectTypeCustomizationsReply(customizations []structs.ObjectTypeCustomization) (structs.GetObjectTypeCustomizationsReply, *structs2.OrionError) {
	var reply = structs.GetObjectTypeCustomizationsReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(customizations) > 0 {
		reply.Header.Success = true
		reply.ObjectTypeCustomizations = customizations
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No customizations were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetObjectTypeCustomizationsAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	reply, myErr := action.getCustomizations(ctx, action.request)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &action.request
	}

	return reply, &action.request
}

func (action GetObjectTypeCustomizationsAction) getCustomizations(ctx context.Context, request structs.GetObjectTypeCustomizationsRequest) (structs.GetObjectTypeCustomizationsReply, *structs2.OrionError) {
	categories, myErr := action.getCustomizationsFromDb(ctx, request)

	if myErr != nil {
		return structs.GetObjectTypeCustomizationsReply{}, myErr
	}

	return action.createGetObjectTypeCustomizationsReply(categories)
}

func (action GetObjectTypeCustomizationsAction) getCustomizationsFromDb(ctx context.Context, request structs.GetObjectTypeCustomizationsRequest) ([]structs.ObjectTypeCustomization, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("object_type_customizations").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs.ObjectTypeCustomization
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
