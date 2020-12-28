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

const SqlGetAllModuleInstallationInfos = "SELECT version, readme, (extract(epoch from installation_date)::bigint)*1000 AS installation_date, " +
	"installed_by, installation_server, module_name FROM releases "

type GetModuleReadmesAction struct {
	baseAction      micro.BaseAction
	MetricsStore    *utils.MetricsStore
	receivedRequest structs.GetModuleReadmesRequest
}

func (action *GetModuleReadmesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetModuleReadmesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	action.receivedRequest = dummy
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetModuleReadmesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetModuleReadmesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetModuleReadmesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetModuleReadmesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetModuleReadmesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetModuleReadmesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetModuleReadmesAction) SendEvents(request micro.IRequest) {

}

func (action GetModuleReadmesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/readme/get"
	var error = "orion/server/misc/error/readme/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetModuleReadmesRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetModuleReadmesReply{})
	info := micro.ActionInformation{
		Name:           "GetModuleReadmesAction",
		Description:    "Gets all readmes of installed modules from the database",
		RequestPath:    "orion/server/misc/request/readme/get",
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

func (action *GetModuleReadmesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetModuleReadmesAction) createGetModuleReadmesReply(objects []structs.ModuleInstallationInfo) (structs.GetModuleReadmesReply, *micro.Exception) {
	var reply = structs.GetModuleReadmesReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Header.Timestamp = utils2.GetCurrentTimeStamp()
	if len(objects) > 0 {
		reply.Header.Success = true
		reply.ModuleInstallationInfos = objects
		return reply, nil
	}
	reply.Header.Success = false
	errorMsg := "No objects were found"
	reply.Header.ErrorMessage = &errorMsg

	err := errors.New(errorMsg)

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetModuleReadmesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.receivedRequest)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.receivedRequest
	}

	reply, myErr := action.getModuleInstallationInfos(action.receivedRequest)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.receivedRequest
	}

	return reply, &action.receivedRequest
}

func (action GetModuleReadmesAction) fillModuleInstallationInfos(rows *sql.Rows) ([]structs.ModuleInstallationInfo, *micro.Exception) {
	var objects []structs.ModuleInstallationInfo

	for rows.Next() {
		var object structs.ModuleInstallationInfo
		/*
			SELECT version, readme, (extract(epoch from installation_date)::bigint)*1000 AS installation_date, installed_by, installation_server, module_name FROM releases
		*/
		err := rows.Scan(&object.Version, &object.Readme, &object.InstallationDate, &object.InstalledBy,
			&object.InstallationServer, &object.ModuleName)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		objects = append(objects, object)
	}
	//fmt.Sprintf("Size of objects: %d", len(objects))
	return objects, nil
}

func (action GetModuleReadmesAction) getModuleInstallationInfos(request structs.GetModuleReadmesRequest) (structs.GetModuleReadmesReply, *micro.Exception) {
	objects, myErr := action.getModuleInstallationInfosFromDb(request)

	if myErr != nil {
		return structs.GetModuleReadmesReply{}, myErr
	}

	return action.createGetModuleReadmesReply(objects)
}

func (action GetModuleReadmesAction) getModuleInstallationInfosFromDb(request structs.GetModuleReadmesRequest) ([]structs.ModuleInstallationInfo, *micro.Exception) {
	var sql = SqlGetAllModuleInstallationInfos

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		sql += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetModuleReadmesAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": sql,
	}).Debug("Issuing GetModuleReadmesAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(sql)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	objects, myErr := action.fillModuleInstallationInfos(rows)
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
