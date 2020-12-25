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
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
	"orion.misc/structs"
	"time"
)

const SqlGetAllStateTransitionRules = "SELECT id, name, description, active, " +
	"(extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, " +
	"from_state, to_states, version FROM state_transition_rules"

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
		Name:           "GetStateTransitionRulesAction",
		Description:    "Gets all state transition rules from the database",
		RequestPath:    "orion/server/misc/request/statetransitionrule/get",
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

func (action *GetStateTransitionRulesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetStateTransitionRulesAction) createGetStateTransitionRulesReply(objects []structs.StateTransitionRule) (structs.GetStateTransitionRulesReply, *micro.Exception) {
	var reply = structs.GetStateTransitionRulesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
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

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetStateTransitionRulesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.receivedRequest
	}

	reply, myErr := action.getStateTransitionRules(action.receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.receivedRequest
	}

	return reply, &action.receivedRequest
}

func (action GetStateTransitionRulesAction) fillStateTransitionRules(rows *sql.Rows) ([]structs.StateTransitionRule, *micro.Exception) {
	var objects []structs.StateTransitionRule

	for rows.Next() {
		var object structs.StateTransitionRule
		/*
			id, name, description, active, " +
			"(extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, " +
			"from_state, to_states, version
		*/
		err := rows.Scan(&object.Info.Id, &object.Info.Name, &object.Info.Description, &object.Info.Active,
			&object.Info.CreatedDate, &object.Info.Alias, &object.FromState, pq.Array(&object.ToStates), &object.Info.Version)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		objects = append(objects, object)
	}
	//fmt.Sprintf("Size of objects: %d", len(objects))
	return objects, nil
}

func (action GetStateTransitionRulesAction) getStateTransitionRules(request structs.GetStateTransitionRulesRequest) (structs.GetStateTransitionRulesReply, *micro.Exception) {
	objects, myErr := action.getStateTransitionRulesFromDb(request)

	if myErr != nil {
		return structs.GetStateTransitionRulesReply{}, myErr
	}

	return action.createGetStateTransitionRulesReply(objects)
}

func (action GetStateTransitionRulesAction) getStateTransitionRulesFromDb(request structs.GetStateTransitionRulesRequest) ([]structs.StateTransitionRule, *micro.Exception) {
	var sql = SqlGetAllStateTransitionRules

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetStateTransitionRulesAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetStateTransitionRulesAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	objects, myErr := action.fillStateTransitionRules(rows)
	if myErr != nil {
		return objects, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return objects, nil
}
