package actions

import (
	"context"
	"database/sql"
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

type GetParametersAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetParametersAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetParametersRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetParametersAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetParametersAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetParametersAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetParametersAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetParametersAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetParametersAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetParametersAction) SendEvents(request micro.IRequest) {

}

func (action GetParametersAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/parameter/get"
	var error = "orion/server/misc/error/parameter/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetParametersRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetParametersReply{})
	info := micro.ActionInformation{
		Name:           "GetParametersAction",
		Description:    "Get attribute definitions based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/parameter/get",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.baseAction.ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *GetParametersAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetParametersAction) createGetParametersReply(parameters []structs.Parameter) (structs.GetParametersReply, *structs2.OrionError) {
	var reply = structs.GetParametersReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(parameters) > 0 {
		reply.Header.Success = true
		reply.Parameters = parameters
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No parameters were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, structs2.NewOrionError(structs2.NoDataFound, err)
}

func (action GetParametersAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetParametersRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	reply, myErr := action.getParameters(ctx, receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetParametersAction) getParameters(ctx context.Context, request structs.GetParametersRequest) (structs.GetParametersReply, *structs2.OrionError) {
	parameters, myErr := action.getParametersFromDb(ctx, request)

	if myErr != nil {
		return structs.GetParametersReply{}, myErr
	}

	return action.createGetParametersReply(parameters)
}

func (action GetParametersAction) getParametersFromDb(ctx context.Context, request structs.GetParametersRequest) ([]structs.Parameter, *structs2.OrionError) {
	cursor, err := action.baseAction.Environment.MongoDbConnection.Database().Collection("parameters").Find(ctx, bson.M{})
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	var objects []structs.Parameter
	if err = cursor.All(ctx, &objects); err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}

	return objects, nil
}
