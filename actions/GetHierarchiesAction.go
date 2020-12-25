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

type GetHierarchiesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

const SqlGetAllHierarchies = "SELECT a.id, name, description, active, (extract(epoch from a.created_date)::bigint)*1000 AS created_date, pretty_id," +
	" b.action_by, c.index, c.object_type" +
	" FROM hierarchies a left outer join cache b on a.id=b.object_id" +
	" LEFT OUTER JOIN ref_hierarchies_types c on a.id=c.hierarchy_id"

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
		Name:           "GetHierarchiesAction",
		Description:    "Get hierarchies based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/hierarchy/get",
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

func (action *GetHierarchiesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetHierarchiesAction) createGetHierarchiesReply(hierarchies []structs.Hierarchy) (structs.GetHierarchiesReply, *micro.Exception) {
	var reply = structs.GetHierarchiesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
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

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetHierarchiesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	var receivedRequest = structs.GetHierarchiesRequest{}

	err := json.Unmarshal(request, &receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	reply, myErr := action.getHierarchies(receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &receivedRequest
	}

	return reply, &receivedRequest
}

func (action GetHierarchiesAction) fillHierarchies(rows *sql.Rows) ([]structs.Hierarchy, *micro.Exception) {
	var hierarchyMap = make(map[int64]structs.Hierarchy)
	for rows.Next() {
		dummy := structs.NewHierarchy()
		var entry structs.HierarchyEntry
		/*
				SELECT a.id, name, description, active, (extract(epoch from a.created_date)::bigint)*1000 AS created_date, pretty_id,
				 b.action_by, c.index, c.object_type
				 FROM hierarchies a left outer join cache b on a.id=b.object_id
			        LEFT OUTER JOIN ref_hierarchies_types c on a.id=c.hierarchy_id
		*/
		err := rows.Scan(&dummy.Info.Id, &dummy.Info.Name, &dummy.Info.Description, &dummy.Info.Active,
			&dummy.Info.CreatedDate, &dummy.Info.Alias, &dummy.Info.LockedBy, &entry.Index, &entry.ObjectType)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		hierarchy, found := hierarchyMap[dummy.Info.Id]
		if found {
			hierarchy.Entries = append(hierarchy.Entries, entry)
			hierarchyMap[hierarchy.Info.Id] = hierarchy
		} else {
			dummy.Entries = append(dummy.Entries, entry)
			hierarchyMap[dummy.Info.Id] = dummy
		}
	}
	hierarchies := make([]structs.Hierarchy, 0, len(hierarchyMap))

	for _, value := range hierarchyMap {
		hierarchies = append(hierarchies, value)
	}
	//fmt.Sprintf("Size of hierarchies: %d", len(hierarchies))
	return hierarchies, nil
}

func (action GetHierarchiesAction) getHierarchies(request structs.GetHierarchiesRequest) (structs.GetHierarchiesReply, *micro.Exception) {
	hierarchies, myErr := action.getHierarchiesFromDb(request)

	if myErr != nil {
		return structs.GetHierarchiesReply{}, myErr
	}

	return action.createGetHierarchiesReply(hierarchies)
}

func (action GetHierarchiesAction) getHierarchiesFromDb(request structs.GetHierarchiesRequest) ([]structs.Hierarchy, *micro.Exception) {
	var sql = SqlGetAllHierarchies

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	sql += " ORDER BY c.index ASC "

	logger := logging.GetLogger("GetHierarchiesAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetHierarchiesAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	states, myErr := action.fillHierarchies(rows)
	if myErr != nil {
		return states, myErr
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("error code: %v - %v", structs2.DatabaseError, err)
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	return states, nil
}
