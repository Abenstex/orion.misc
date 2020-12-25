package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mqtt"
	"github.com/abenstex/orion.commons/app"
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

type SaveObjectCategoryReferenceAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	setRequest   structs2.SaveObjectCategoryReferenceRequest
}

func (action *SaveObjectCategoryReferenceAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveObjectCategoryReferenceRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, action, true)
	if err != nil {
		return micro.NewException(structs.RequestHeaderInvalid, err)
	}
	action.setRequest = dummy

	return nil
}

func (action *SaveObjectCategoryReferenceAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action *SaveObjectCategoryReferenceAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action *SaveObjectCategoryReferenceAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveObjectCategoryReferenceAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveObjectCategoryReferenceAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveObjectCategoryReferenceAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveObjectCategoryReferenceAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveObjectCategoryReferenceRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveObjectCategoryReferenceAction",
			action.GetBaseAction().Environment,
			true).Warn("Events won't be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs2.CategoryReferencesSavedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Categories: action.getCategoryKeysAsSlice(),
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveObjectCategoryReferenceAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveObjectCategoryReferenceAction) getCategoryKeysAsSlice() []uint64 {
	keys := make([]uint64, 0, len(action.setRequest.CategoryReferences))
	for _, ref := range action.setRequest.CategoryReferences {
		keys = append(keys, ref.CategoryId)
	}

	return keys
}

func (action SaveObjectCategoryReferenceAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/savereferences"
	var error = "orion/server/misc/error/category/savereferences"
	var event = "orion/server/misc/event/category/savereferences"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveObjectCategoryReferenceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.CategoryReferencesSavedEvent{})
	info := micro.ActionInformation{
		Name:           "SaveObjectCategoryReferenceAction",
		Description:    "Saves attribute values and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/category/savereferences",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		EventSample:    dataStructures.JsonNullString{NullString: sql.NullString{String: eventSample, Valid: true}},
		IsScriptable:   true,
	}

	return info
}

func (action *SaveObjectCategoryReferenceAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveObjectCategoryReferenceAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	/*err := json.Unmarshal(request, &action.setRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.setRequest
	}*/

	dummy, _ := json.Marshal(action.setRequest)
	fmt.Printf("Request: %v\n", string(dummy))

	exception := action.saveCategoryReference(action.setRequest)
	if exception != nil {
		logging.GetLogger("SaveObjectCategoryReferenceAction",
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

func (action SaveObjectCategoryReferenceAction) saveSingleCategory(reference structs2.CategoryReference, user string, txn *sql.Tx) *micro.Exception {
	insertQuery := "INSERT INTO ref_categories_objects (category_id, object_type, object_id, action_by, object_version) " +
		" VALUES ($1, $2, $3, $4, $5) " +
		" ON CONFLICT ON CONSTRAINT ref_categories_objects_unique_constraint DO NOTHING "

	_, err := txn.Exec(insertQuery, reference.CategoryId, reference.ObjectType, reference.ObjectId, user, reference.ObjectVersion)
	if err != nil {
		txn.Rollback()
		return micro.NewException(structs.DatabaseError, err)
	}

	return nil
}

func (action SaveObjectCategoryReferenceAction) saveCategoryReference(request structs2.SaveObjectCategoryReferenceRequest) *micro.Exception {
	txn, err := action.baseAction.Environment.Database.Begin()
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).
			WithError(err).
			Error("transaction could not be started")
		return micro.NewException(structs.DatabaseError, err)
	}
	catIds := action.getCategoryKeysAsSlice()
	deleteQuery := "DELETE FROM ref_categories_objects WHERE category_id = ANY($1::bigint[])"

	_, err = txn.Exec(deleteQuery, pq.Array(catIds))
	if err != nil {
		txn.Rollback()
		return micro.NewException(structs.DatabaseError, err)
	}

	for _, reference := range request.CategoryReferences {
		exception := action.saveSingleCategory(reference, request.Header.User, txn)
		if exception != nil {
			return exception
		}
	}
	err = txn.Commit()

	return micro.NewException(structs.DatabaseError, err)
}
