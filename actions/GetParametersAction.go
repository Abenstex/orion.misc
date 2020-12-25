package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	utils2 "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"orion.misc/structs"
	"time"
)

const SqlGetAllParameters = "SELECT id, name, description, active," +
	" (extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, b.action_by, a.value" +
	" FROM parameters a left outer join cache b on a.id=b.object_id "

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

func (action GetParametersAction) createGetParametersReply(parameters []structs.Parameter) (structs.GetParametersReply, *micro.Exception) {
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

	return reply, micro.NewException(structs2.NoDataFound, err)
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

	reply, myErr := action.getParameters(receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetParametersAction) fillParameters(rows *sql.Rows) ([]structs.Parameter, *micro.Exception) {
	var parameters []structs.Parameter

	for rows.Next() {
		var attribute structs.Parameter
		/*
			id, name, description, active, " +
			"(extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, b.action_by, datatype, overwriteable, " +
			"allowed_object_types, list_of_values, numeric_from, numeric_to, query, default_value, assign_during_object_creation
		*/
		err := rows.Scan(&attribute.Info.Id, &attribute.Info.Name, &attribute.Info.Description, &attribute.Info.Active,
			&attribute.Info.CreatedDate, &attribute.Info.Alias, &attribute.Info.LockedBy, &attribute.Value)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		parameters = append(parameters, attribute)
	}
	//fmt.Sprintf("Size of parameters: %d", len(parameters))
	return parameters, nil
}

func (action GetParametersAction) getParameters(request structs.GetParametersRequest) (structs.GetParametersReply, *micro.Exception) {
	parameters, myErr := action.getParametersFromDb(request)

	if myErr != nil {
		return structs.GetParametersReply{}, myErr
	}

	return action.createGetParametersReply(parameters)
}

func (action GetParametersAction) getParametersFromDb(request structs.GetParametersRequest) ([]structs.Parameter, *micro.Exception) {
	var sql = SqlGetAllParameters

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetParametersAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetParametersAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	parameters, myErr := action.fillParameters(rows)
	if myErr != nil {
		return parameters, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return parameters, nil
}
