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

const SqlGetAllObjectTypeCustomizations = "SELECT id, object_type, field_name, field_data_type, " +
	" (extract(epoch from created_date)::bigint)*1000 AS created_date, field_mandatory, field_default_value, created_by " +
	" FROM object_type_customizations "

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

	return micro.NewException(structs2.RequestHeaderInvalid, err)
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
		Name:           "GetObjectTypeCustomizationsAction",
		Description:    "Get object type customizations based on conditions or all if no conditions were sent in the request",
		RequestPath:    "orion/server/misc/request/objectcustomization/get",
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

func (action *GetObjectTypeCustomizationsAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action GetObjectTypeCustomizationsAction) createGetObjectTypeCustomizationsReply(customizations []structs.ObjectTypeCustomization) (structs.GetObjectTypeCustomizationsReply, *micro.Exception) {
	var reply = structs.GetObjectTypeCustomizationsReply{}
	reply.Header = structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
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

	return reply, micro.NewException(structs2.NoDataFound, err)
}

func (action GetObjectTypeCustomizationsAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	reply, myErr := action.getCustomizations(action.request)
	if myErr != nil {
		return structs2.NewErrorReplyHeaderWithException(myErr,
			action.ProvideInformation().ErrorReplyPath.String), &action.request
	}

	return reply, &action.request
}

func (action GetObjectTypeCustomizationsAction) fillObjectTypeCustomizations(rows *sql.Rows) ([]structs.ObjectTypeCustomization, *micro.Exception) {
	var customizations []structs.ObjectTypeCustomization

	for rows.Next() {
		var customization structs.ObjectTypeCustomization

		err := rows.Scan(&customization.Id, &customization.ObjectType, &customization.FieldName, &customization.FielDataType,
			&customization.CreatedDate, &customization.FieldMandatory, &customization.FieldDefaultValue, &customization.CreatedBy)
		if err != nil {
			return nil, micro.NewException(structs2.DatabaseError, err)
		}

		customizations = append(customizations, customization)
	}

	return customizations, nil
}

func (action GetObjectTypeCustomizationsAction) getCustomizations(request structs.GetObjectTypeCustomizationsRequest) (structs.GetObjectTypeCustomizationsReply, *micro.Exception) {
	categories, myErr := action.getCustomizationsFromDb(request)

	if myErr != nil {
		return structs.GetObjectTypeCustomizationsReply{}, myErr
	}

	return action.createGetObjectTypeCustomizationsReply(categories)
}

func (action GetObjectTypeCustomizationsAction) getCustomizationsFromDb(request structs.GetObjectTypeCustomizationsRequest) ([]structs.ObjectTypeCustomization, *micro.Exception) {
	var query = SqlGetAllObjectTypeCustomizations

	if request.WhereClause != nil && len(*request.WhereClause) > 1 {
		query += " WHERE " + *request.WhereClause
	}
	logger := logging.GetLogger("GetObjectTypeCustomizationsAction", action.GetBaseAction().Environment, false)
	logger.WithFields(logrus.Fields{
		"query": query,
	}).Debug("Issuing GetObjectTypeCustomizationsAction query")

	rows, err := action.GetBaseAction().Environment.Database.Query(query)
	if err != nil {
		return nil, micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()
	categories, myErr := action.fillObjectTypeCustomizations(rows)
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
