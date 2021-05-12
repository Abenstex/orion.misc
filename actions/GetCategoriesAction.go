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

type GetCategoriesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	request      structs.GetCategoriesRequest
}

func (action GetCategoriesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetCategoriesRequest{}
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

func (action GetCategoriesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetCategoriesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetCategoriesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetCategoriesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetCategoriesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetCategoriesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetCategoriesAction) SendEvents(request micro.IRequest) {

}

func (action GetCategoriesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/get"
	var error = "orion/server/misc/error/category/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetCategoriesRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetCategoriesReply{})
	info := micro.ActionInformation{
		Name:            "GetCategoriesAction",
		Description:     "Get categories based on conditions or all if no conditions were sent in the request",
		RequestTopic:    "orion/server/misc/request/category/get",
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

func (action *GetCategoriesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetCategoriesAction) createGetCategoriesReply(categories []structs.Category) (structs.GetCategoriesReply, *structs2.OrionError) {
	var reply = structs.GetCategoriesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(categories) > 0 {
		reply.Header.Success = true
		reply.Categories = categories
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No categories were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetCategoriesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	reply, myErr := action.getCategories(ctx, action.request)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyTopic), &action.request
	}

	return reply, &action.request
}

func (action GetCategoriesAction) getCategories(ctx context.Context, request structs.GetCategoriesRequest) (structs.GetCategoriesReply, *structs2.OrionError) {
	categories, myErr := action.getCategoriesFromDb(ctx, request)

	if myErr != nil {
		return structs.GetCategoriesReply{}, myErr
	}

	return action.createGetCategoriesReply(categories)
}

func (action GetCategoriesAction) getCategoriesFromDb(ctx context.Context, request structs.GetCategoriesRequest) ([]structs.Category, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("categories").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs.Category
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
