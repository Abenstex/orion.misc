package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mqtt"
	laniakea "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	"github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"net/http"
	structs2 "orion.misc/structs"
	"time"
)

type SaveObjectTypeCustomizationAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.ObjectTypeCustomization
	saveRequest  structs2.SaveObjectTypeCustomizationsRequest
}

func (action *SaveObjectTypeCustomizationAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveObjectTypeCustomizationsRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)
	action.saveRequest = dummy

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action SaveObjectTypeCustomizationAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveObjectTypeCustomizationAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveObjectTypeCustomizationAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveObjectTypeCustomizationAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveObjectTypeCustomizationAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveObjectTypeCustomizationAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveObjectTypeCustomizationAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveObjectTypeCustomizationsRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveObjectTypeCustomizationAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.ObjectTypeCustomizations))
	for _, parameter := range saveRequest.ObjectTypeCustomizations {
		ids = append(ids, parameter.Id.Int64)
	}
	event := structs2.ObjectTypeCustomizationsSavedEvent{
		Header:                   *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		ObjectTypeCustomizations: ids,
		ObjectType:               "OBJECT_TYPE_CUSTOMIZATION",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveObjectTypeCustomizationAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveObjectTypeCustomizationAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/objectcustomization/save"
	var error = "orion/server/misc/error/objectcustomization/save"
	var event = "orion/server/misc/event/objectcustomization/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveObjectTypeCustomizationsRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.ObjectTypeCustomizationsSavedEvent{})
	info := micro.ActionInformation{
		Name:           "SaveObjectTypeCustomizationAction",
		Description:    "Saves Object Type Customizations and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/objectcustomization/save",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		EventSample:    dataStructures.JsonNullString{NullString: sql.NullString{String: eventSample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *SaveObjectTypeCustomizationAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveObjectTypeCustomizationAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	exception := action.saveObjects(action.saveRequest.ObjectTypeCustomizations, action.saveRequest.Header.User)
	if exception != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SaveObjectTypeCustomizationAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithException(exception,
			action.ProvideInformation().ErrorReplyPath.String), &action.saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.saveRequest
}

func (action *SaveObjectTypeCustomizationAction) saveObjects(customizations []structs2.ObjectTypeCustomization, user string) *micro.Exception {
	insertSql := "INSERT INTO object_type_customizations (object_type, field_name, " +
		"field_data_type, field_mandatory, field_default_value, " +
		"created_by) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	updateSql := "UPDATE object_type_customizations SET  object_type = $1, field_name = $2, " +
		"field_data_type = $3, field_mandatory = $4, field_default_value = $5 " +
		" WHERE id = $6"
	action.savedObjects = make([]structs2.ObjectTypeCustomization, len(customizations), len(customizations))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return micro.NewException(structs.DatabaseError, err)
	}
	for idx, customization := range customizations {
		var id int64
		if !customization.Id.Valid || customization.Id.Int64 <= 0 {
			err = laniakea.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, customization.ObjectType,
				customization.FieldName, customization.FielDataType, customization.FieldMandatory,
				customization.FieldDefaultValue, user)
			if err != nil {
				logging.GetLogger("SaveObjectTypeCustomizationAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not insert customization")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
			customization.Id = dataStructures.JsonNullInt64{NullInt64: sql.NullInt64{
				Int64: id,
				Valid: true,
			}}
		} else {
			err := laniakea.ExecuteQueryWithTransaction(txn, updateSql, customization.ObjectType,
				customization.FieldName, customization.FielDataType, customization.FieldMandatory,
				customization.FieldDefaultValue, customization.Id.Int64)
			if err != nil {
				logging.GetLogger("SaveParametersAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not update customization")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
		}

		action.savedObjects[idx] = customization
	}
	err = txn.Commit()
	if err != nil {
		return micro.NewException(structs.DatabaseError, err)
	}

	return nil
}
