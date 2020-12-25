package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mqtt"
	utils2 "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	"github.com/abenstex/orion.commons/couchdb"
	http2 "github.com/abenstex/orion.commons/http"
	"github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	structs2 "orion.misc/structs"
	"time"
)

type SetAttributeValueAction struct {
	baseAction       micro.BaseAction
	MetricsStore     *utils.MetricsStore
	setRequest       structs2.SetAttributeValueRequest
	attributeChanges []structs2.AttributeChange
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
	action.setRequest = dummy

	return action.getOldValueBeforeUpdate(dummy)
}

func (action *SetAttributeValueAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *SetAttributeValueAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *SetAttributeValueAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {
	requestString, _ := action.setRequest.ToString()

	// HistoricizeAttributeChange(request string, requestPath, oldValue, newValue, referencedType string, receivedTime int64)
	for _, change := range action.attributeChanges {
		err := couchdb.HistoricizeAttributeChange(requestString, action.ProvideInformation().RequestPath, change.OriginalValue,
			change.NewValue, change.ObjectType, change.AttributeId, change.ObjectId, action.setRequest.Header.ReceivedTimeInMillis, change.ObjectVersion)
		if err != nil {
			logging.GetLogger("SetAttributeValueAction",
				action.GetBaseAction().Environment,
				true).WithError(err).Error("cannot historicize attribute value change")

			break
		}
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
		Header:           *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		AttributeChanges: action.attributeChanges,
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

	exception := action.saveAttribute(action.setRequest)
	if exception != nil {
		logging.GetLogger("SetAttributeValueAction",
			action.GetBaseAction().Environment,
			true).WithFields(logrus.Fields{
			"code":  exception.ErrorCode,
			"error": exception.ErrorText,
		}).Error("Data could not be saved")

		return structs.NewErrorReplyHeaderWithException(exception,
			action.ProvideInformation().ErrorReplyPath.String), &action.setRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.setRequest
}

func (action SetAttributeValueAction) saveAttribute(request structs2.SetAttributeValueRequest) *micro.Exception {
	query := "INSERT INTO ref_attributes_objects (attr_id, attr_value, object_type, object_id, action_by, object_version) " +
		"VALUES ($1, $2, $3, $4, $5, $6) " +
		"ON CONFLICT ON CONSTRAINT ref_attributes_objects_unique_constraint " +
		"DO UPDATE SET attr_value = $7, action_by = $8 "

	txn, err := action.baseAction.Environment.Database.Begin()
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).
			WithError(err).
			Error("transaction could not be started")
		return micro.NewException(structs.DatabaseError, err)
	}

	for _, attribute := range request.Attributes {
		err = utils2.ExecuteQueryWithTransaction(txn, query, attribute.Info.Id, attribute.Value,
			attribute.ObjectType, attribute.ObjectId, request.Header.User, attribute.ObjectVersion, attribute.Value, request.Header.User)
		if err != nil {
			txn.Rollback()
			logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).
				WithError(err).
				Error("values could not be saved -> rollback")
			return micro.NewException(structs.DatabaseError, err)
		}
	}
	err = txn.Commit()

	return micro.NewException(structs.DatabaseError, err)
}

func (action *SetAttributeValueAction) getOldValueBeforeUpdate(request structs2.SetAttributeValueRequest) *micro.Exception {
	var attrIds []int64
	var objectIds []uint64
	newValueMap := make(map[uint64]structs.Attribute, len(request.Attributes))
	for _, tmp := range request.Attributes {
		attrIds = append(attrIds, tmp.Info.Id)
		objectIds = append(objectIds, tmp.ObjectId)
		newValueMap[uint64(tmp.Info.Id)] = tmp
	}
	query := "SELECT attr_value, attr_id, object_id, object_type, object_version " +
		"FROM ref_attributes_objects WHERE attr_id = ANY($1::bigint[]) AND object_id = ANY($2::bigint[])"
	rows, err := action.GetBaseAction().Environment.Database.Query(query, pq.Array(attrIds), pq.Array(objectIds))
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).
			WithError(err).
			Error("original values could not be read from database")
		return micro.NewException(structs.DatabaseError, err)
	}
	defer rows.Close()

	var attributeChanges []structs2.AttributeChange
	for rows.Next() {
		var change structs2.AttributeChange
		err := rows.Scan(&change.OriginalValue, &change.AttributeId, &change.ObjectId, &change.ObjectType, &change.ObjectVersion)
		if err != nil {
			logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).
				WithError(err).
				Error("original values could not be read from database")
			return micro.NewException(structs.DatabaseError, err)
		}
		change.NewValue = newValueMap[change.AttributeId].Value
		attributeChanges = append(attributeChanges, change)
	}

	action.attributeChanges = attributeChanges

	return nil
}
