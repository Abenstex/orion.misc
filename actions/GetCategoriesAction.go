package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"laniakea/dataStructures"
	"laniakea/logging"
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

const SqlGetAllCategories = "SELECT id, name, description, active," +
	" (extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, b.action_by, a.referenced_type" +
	" FROM categories a left outer join cache b on a.id=b.object_id "

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

	return micro.NewException(structs2.RequestHeaderInvalid, err)
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
		Name:           "GetCategoriesAction",
		Description:    "Get categories based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/category/get",
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

func (action *GetCategoriesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetCategoriesAction) createGetCategoriesReply(categories []structs.Category) (structs.GetCategoriesReply, *micro.Exception) {
	var reply = structs.GetCategoriesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
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

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetCategoriesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	reply, myErr := action.getCategories(action.request)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.request
	}

	return reply, &action.request
}

func (action GetCategoriesAction) fillCategories(rows *sql.Rows) ([]structs.Category, *micro.Exception) {
	var categories []structs.Category

	for rows.Next() {
		var category structs.Category

		err := rows.Scan(&category.Info.Id, &category.Info.Name, &category.Info.Description, &category.Info.Active,
			&category.Info.CreatedDate, &category.Info.Alias, &category.Info.LockedBy, &category.ReferencedType)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (action GetCategoriesAction) getCategories(request structs.GetCategoriesRequest) (structs.GetCategoriesReply, *micro.Exception) {
	categories, myErr := action.getCategoriesFromDb(request)

	if myErr != nil {
		return structs.GetCategoriesReply{}, myErr
	}

	return action.createGetCategoriesReply(categories)
}

func (action GetCategoriesAction) getCategoriesFromDb(request structs.GetCategoriesRequest) ([]structs.Category, *micro.Exception) {
	var query = SqlGetAllCategories

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		query += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetCategoriesAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": query,
	}).Debug("Issuing GetCategoriesAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(query)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	categories, myErr := action.fillCategories(rows)
	if myErr != nil {
		return categories, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return categories, nil
}
