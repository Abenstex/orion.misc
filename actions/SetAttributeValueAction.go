package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/spf13/viper"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	"laniakea/mqtt"
	utils2 "laniakea/utils"
	"net/http"
	"orion.commons/app"
	"orion.commons/couchdb"
	http2 "orion.commons/http"
	"orion.commons/structs"
	"orion.commons/utils"
	structs2 "orion.misc/structs"
	"time"
)

type SetAttributeValueAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	setRequest    structs2.SetAttributeValueRequest
	originalValue string
}

func (action *SetAttributeValueAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SetAttributeValueRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}
	err = action.getOldValueBeforeUpdate(dummy.AttributeId, dummy.ObjectId)

	return micro.NewException(structs.DatabaseError, err)
}

func (action *SetAttributeValueAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *SetAttributeValueAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *SetAttributeValueAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {
	requestString, _ := action.setRequest.ToString()

	// HistoricizeAttributeChange(request string, requestPath, oldValue, newValue, referencedType string, receivedTime int64)
	err := couchdb.HistoricizeAttributeChange(requestString, action.ProvideInformation().RequestPath, action.originalValue,
		action.setRequest.Value, action.setRequest.ObjectType, action.setRequest.AttributeId, action.setRequest.ObjectId, action.setRequest.Header.ReceivedTimeInMillis)
	if err != nil {
		logging.GetLogger("SetAttributeValueAction",
			action.GetBaseAction().Environment,
			true).WithError(err).Error("cannot historicize attribute value change")
	}
}

func (action SetAttributeValueAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SetAttributeValueAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SetAttributeValueAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SetAttributeValueAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SetAttributeValueRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SetAttributeValueAction",
			action.GetBaseAction().Environment,
			true).Warn("Events won't be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs2.AttributeValueChangedEvent{
		Header:      *micro.NewEventHeaderForAction(action.ProvideInformation(), request.GetHeader().SenderId, ""),
		ObjectType:  action.setRequest.ObjectType,
		ObjectId:    action.setRequest.ObjectId,
		Value:       action.setRequest.Value,
		AttributeId: action.setRequest.AttributeId,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SetAttributeValueAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SetAttributeValueAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attribute/set"
	var error = "orion/server/misc/error/attribute/set"
	var event = "orion/server/misc/event/attribute/set"
	var requestSample = dataStructures.StructToJsonString(structs2.DefineAttributeRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "SetAttributeValueAction",
		Description:    "Saves attribute values and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/attribute/set",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   true,
	}

	return info
}

func (action *SetAttributeValueAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SetAttributeValueAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	err := json.Unmarshal(request, &action.setRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.setRequest
	}

	err = action.saveAttribute(action.setRequest)
	if err != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SetAttributeValueAction",
			action.GetBaseAction().Environment,
			true).WithError(err).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.setRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.setRequest
}

func (action SetAttributeValueAction) saveAttribute(request structs2.SetAttributeValueRequest) error {
	query := "INSERT INTO ref_attributes_objects (attr_id, attr_value, object_type, object_id, action_by) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"ON CONFLICT ON CONSTRAINT ref_attributes_objects_unique_constraint " +
		"DO UPDATE SET attr_value = $6, action_by = $7 "

	return utils2.ExecuteQueryInTransaction(action.baseAction.Environment, query, request.AttributeId, request.Value,
		request.ObjectType, request.ObjectId, request.Header.User, request.Value, request.Header.User)
}

func (action *SetAttributeValueAction) getOldValueBeforeUpdate(attr_id, object_id uint64) error {
	query := "SELECT attr_value FROM ref_attributes_objects WHERE attr_id=$1 AND object_id=$2"
	var value string
	row := action.GetBaseAction().Environment.Database.QueryRow(query, attr_id, object_id)
	err := row.Scan(&value)

	action.originalValue = value

	return err
}
