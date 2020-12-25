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
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"net/http"
	structs2 "orion.misc/structs"
	"time"
)

type DefineAttributesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs.AttributeDefinition
}

func (action DefineAttributesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.DefineAttributeRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action DefineAttributesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action DefineAttributesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action DefineAttributesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action DefineAttributesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DefineAttributesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *DefineAttributesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DefineAttributesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.DefineAttributeRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DefineAttributesAction",
			action.GetBaseAction().Environment,
			true).Warn("Events won't be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs2.AttributeDefinitionSavedEvent{
		Header:               *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		AttributeDefinitions: action.savedObjects,
		ObjectType:           "AttributeDefinition",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DefineAttributesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/save"
	var error = "orion/server/misc/error/attributedefinition/save"
	var event = "orion/server/misc/event/attributedefinition/save"
	var requestSample = dataStructures.StructToJsonString(structs2.DefineAttributeRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "DefineAttributesAction",
		Description:    "Saves AttributeDefinition and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/attributedefinition/save",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *DefineAttributesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DefineAttributesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs2.DefineAttributeRequest{}

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	err = action.saveObjects(saveRequest.UpdatedAttributeDefinitions, saveRequest.OriginalAttributeDefintions, saveRequest.Header.User)
	if err != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("DefineAttributesAction",
			action.GetBaseAction().Environment,
			true).WithError(err).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &saveRequest
}

func (action *DefineAttributesAction) saveObjects(updatedObjects []structs.AttributeDefinition, originalObjects []structs.AttributeDefinition, user string) error {
	insertSql := "INSERT INTO attributes (name, description, active, action_by, pretty_id, datatype, " +
		" overwriteable, allowed_object_types, list_of_values, numeric_from, numeric_to, " +
		" query, object_type, default_value, assign_during_object_creation) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id"
	updateSql := "UPDATE attributes SET name = $1, description = $2, active = $3, action_by = $4, " +
		"pretty_id = $5, overwriteable = $6, allowed_object_types = $7, list_of_values = $8, " +
		"numeric_from = $9, numeric_to = $10, query = $11, datatype = $12, default_value=$13, assign_during_object_creation = $14 WHERE id = $15 "
	action.savedObjects = make([]structs.AttributeDefinition, len(updatedObjects), len(updatedObjects))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return err
	}
	for idx, updatedObject := range updatedObjects {
		var id int64
		if updatedObject.Info.Id <= 0 {
			err = laniakea.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, updatedObject.Info.Name,
				updatedObject.Info.Description, updatedObject.Info.Active, user,
				updatedObject.Info.Alias, updatedObject.DataType, updatedObject.Overwriteable, pq.Array(updatedObject.AllowedObjectTypes),
				pq.Array(updatedObject.ListOfValues), updatedObject.NumericFrom, updatedObject.NumericTo,
				updatedObject.Query, "ATTRIBUTE", updatedObject.DefaultValue, updatedObject.AssignDuringObjectCreation)
			if err != nil {
				logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not insert user")
				txn.Rollback()
				return err
			}
			updatedObject.Info.Id = id
		} else {
			err := laniakea.ExecuteQueryWithTransaction(txn, updateSql, updatedObject.Info.Name,
				updatedObject.Info.Description, updatedObject.Info.Active, user, updatedObject.Info.Alias, updatedObject.Overwriteable,
				pq.Array(updatedObject.AllowedObjectTypes), pq.Array(updatedObject.ListOfValues),
				updatedObject.NumericFrom, updatedObject.NumericTo, updatedObject.Query, updatedObject.DataType,
				updatedObject.DefaultValue, updatedObject.AssignDuringObjectCreation, updatedObject.Info.Id)
			if err != nil {
				logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not update user")
				txn.Rollback()
				return err
			}
		}

		action.savedObjects[idx] = updatedObject
	}

	return txn.Commit()
}
