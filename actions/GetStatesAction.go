package actions

import (
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

type GetStatesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
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
	sampleRequest, sampleReply := action.sampleRequestReply()
	requestJson, _ := sampleRequest.ToString()
	replyJson, _ := sampleReply.MarshalJSON()
	info := micro.ActionInformation{
		Name:           "GetStatesAction",
		Description:    "Get states based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/state/get",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.baseAction.ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestJson, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replyJson, Valid: true}},
	}

	return info
}

func (action GetStatesAction) sampleRequestReply() (micro.IRequest, micro.IReply) {
	perm1 := structs.State{Info: structs2.NewBaseInfo()}
	perm2 := structs.State{Info: structs2.NewBaseInfo()}
	whereClause := " id=123 "
	request := structs.GetStatesRequest{
		Header:      micro.SampleRequestHeader(),
		WhereClause: &whereClause,
	}
	reply := structs.GetStatesReply{
		Header: micro.SampleReplyHeader(),
		States: []structs.State{perm1, perm2},
	}

	return &request, reply
}

func (action *GetStatesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetStatesAction) createGetStatesReply(states []structs.State) (structs.GetStatesReply, *structs2.OrionError) {
	var reply = structs.GetStatesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
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

func (action GetStatesAction) HeyHo(request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetStatesRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}
	err = app.DefaultHandleActionRequest(request, &receivedRequest.Header, &action, true)
	if err != nil {
		orionError := structs2.NewOrionError(structs2.GeneralError, err)
		return structs2.NewErrorReplyHeaderWithOrionErr(orionError,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}
	reply, myErr := action.getStates(receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetStatesAction) fillStates(rows *sql.Rows) ([]structs.State, *structs2.OrionError) {
	var states []structs.State

	for rows.Next() {
		var state structs.State
		/*
			id, name, description, active, (extract(epoch from created_date)::bigint)*1000 AS created_date, pretty_id, "+
							" b.action_by, referenced_type, object_available, substate, default_state
		*/
		err := rows.Scan(&state.Info.Id, &state.Info.Name, &state.Info.Description, &state.Info.Active,
			&state.Info.CreatedDate, &state.Info.Alias, &state.Info.LockedBy, &state.ReferencedType,
			&state.ObjectAvailable, &state.Substate, &state.DefaultState)
		if err != nil {
			return nil, structs2.NewOrionError(structs2.DatabaseError, err)
		}

		states = append(states, state)
	}
	//fmt.Sprintf("Size of states: %d", len(states))
	return states, nil
}

func (action GetStatesAction) getStates(request structs.GetStatesRequest) (structs.GetStatesReply, *structs2.OrionError) {
	states, myErr := action.getStatesFromDb(request)

	if myErr != nil {
		return structs.GetStatesReply{}, myErr
	}

	return action.createGetStatesReply(states)
}

func (action GetStatesAction) getStatesFromDb(request structs.GetStatesRequest) ([]structs.State, *structs2.OrionError) {
	var sql = SqlGetAllStates

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetStatesAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetStatesAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	defer rows.Close()
	states, myErr := action.fillStates(rows)
	if myErr != nil {
		return states, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, structs2.NewOrionError(structs2.DatabaseError, err)
	}
	return states, nil
}
