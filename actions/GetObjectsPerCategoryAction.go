package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"laniakea/dataStructures"
	"laniakea/micro"
	utils2 "laniakea/utils"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	structs2 "orion.commons/structs"
	"orion.commons/utils"
	"orion.misc/structs"
	"time"
)

const SqlGetObjectsPerCategory = "SELECT b.object_id, b.object_type, b.object_version " +
	" FROM categories a INNER JOIN ref_categories_objects b ON a.id = b.category_id " +
	" WHERE a.id=$1"

type GetObjectsPerCategoryAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	getRequest   structs.GetObjectsPerCategoriesRequest
}

func (action *GetObjectsPerCategoryAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetObjectsPerCategoriesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	if dummy.CategoryId == nil || *dummy.CategoryId < 0 {
		return micro.NewException(structs2.MissingParameterError, fmt.Errorf("the parameter categoryId is missing in the request"))
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)

	action.getRequest = dummy

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetObjectsPerCategoryAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetObjectsPerCategoryAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetObjectsPerCategoryAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetObjectsPerCategoryAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetObjectsPerCategoryAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetObjectsPerCategoryAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetObjectsPerCategoryAction) SendEvents(request micro.IRequest) {

}

func (action GetObjectsPerCategoryAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/getobjects"
	var error = "orion/server/misc/error/category/getobjects"
	var requestSample = dataStructures.StructToJsonString(structs.GetObjectsPerCategoriesRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetObjectsPerCategoriesReply{})
	info := micro.ActionInformation{
		Name:           "GetObjectsPerCategoryAction",
		Description:    "Get all objects associated to a category (categoryId: mandatory)",
		RequestPath:    "orion/server/misc/request/category/getobjects",
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

func (action *GetObjectsPerCategoryAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetObjectsPerCategoryAction) createGetObjectsPerCategoryReply(objects []structs2.BaseInfo) (structs.GetObjectsPerCategoriesReply, *micro.Exception) {
	var reply = structs.GetObjectsPerCategoriesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(objects) > 0 {
		reply.Header.Success = true
		reply.Objects = objects
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No objects were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetObjectsPerCategoryAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.getRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.getRequest
	}

	reply, myErr := action.getObjectsPerCategory(action.getRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.getRequest
	}

	return reply, &action.getRequest
}

func (action GetObjectsPerCategoryAction) fillBaseInfos(rows *sql.Rows) ([]structs2.BaseInfo, *micro.Exception) {
	var objects []structs2.BaseInfo

	for rows.Next() {
		var object structs2.BaseInfo
		/*
			b.object_id, b.object_type, b.object_version
		*/
		err := rows.Scan(&object.Id, &object.ObjectType, &object.Version)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		objects = append(objects, object)
	}

	return objects, nil
}

func (action GetObjectsPerCategoryAction) getObjectsPerCategory(request structs.GetObjectsPerCategoriesRequest) (structs.GetObjectsPerCategoriesReply, *micro.Exception) {
	objects, myErr := action.getObjectsPerCategoryFromDb(request)

	if myErr != nil {
		return structs.GetObjectsPerCategoriesReply{}, myErr
	}

	return action.createGetObjectsPerCategoryReply(objects)
}

func (action GetObjectsPerCategoryAction) getObjectsPerCategoryFromDb(request structs.GetObjectsPerCategoriesRequest) ([]structs2.BaseInfo, *micro.Exception) {
	rows, err := action.GetBaseAction().Environment.Database.Query(SqlGetObjectsPerCategory, action.getRequest.CategoryId)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()

	return action.fillBaseInfos(rows)
}
