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
	http2 "orion.commons/http"
	structs2 "orion.commons/structs"
	"orion.commons/utils"
	"orion.misc/structs"
	"time"
)

type SaveHierarchiesAction struct {
	baseAction       micro.BaseAction
	MetricsStore     *utils.MetricsStore
	savedHierarchies []structs.Hierarchy
}

func (action SaveHierarchiesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveHierarchiesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Errorf("error unmarshalling the request: %v\n", string(request))
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action SaveHierarchiesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveHierarchiesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveHierarchiesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveHierarchiesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveHierarchiesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveHierarchiesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveHierarchiesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveHierarchiesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveHierarchiesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.UpdatedHierarchies))
	for _, hierarchy := range saveRequest.UpdatedHierarchies {
		ids = append(ids, hierarchy.Info.Id)
	}
	event := structs.SavedHierarchiesEvent{
		Header:      *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Hierarchies: action.savedHierarchies,
		ObjectType:  "HIERARCHY",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveHierarchiesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveHierarchiesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/hierarchy/save"
	var error = "orion/server/misc/error/hierarchy/save"
	var event = "orion/server/misc/event/hierarchy/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs.SavedHierarchiesEvent{})
	info := micro.ActionInformation{
		Name:           "SaveHierarchiesAction",
		Description:    "Saves hierarchies to the database",
		RequestPath:    "orion/server/misc/request/hierarchy/save",
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

func (action *SaveHierarchiesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveHierarchiesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs.SaveHierarchiesRequest{}

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithException(micro.NewException(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	exception := action.saveHierarchies(saveRequest.UpdatedHierarchies, saveRequest.Header.User)
	if exception != nil {
		logging.GetLogger(action.ProvideInformation().Name,
			action.GetBaseAction().Environment,
			true).WithField("exception", exception).Error("Data could not be saved")
		return structs2.NewErrorReplyHeaderWithException(exception,
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	reply := structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveHierarchiesAction) saveHierarchies(updatedHierarchies []structs.Hierarchy, user string) *micro.Exception {
	// INSERT INTO hierarchies (id, name, description, active, action_by, created_date, pretty_id, action_date, object_type, service, index, hierarchy_object_type) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
	insertSql := "INSERT INTO hierarchies (name, description, active, action_by, " +
		"pretty_id, object_type) " +
		"VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

	action.savedHierarchies = make([]structs.Hierarchy, len(updatedHierarchies), len(updatedHierarchies))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return micro.NewException(structs2.DatabaseError, err)
	}

	for idx, hierarchy := range updatedHierarchies {
		var id int64
		if hierarchy.Info.Id > 0 {
			err = utils.DeleteObjectByIdWithTransaction(txn, "ref_hierarchies_types", hierarchy.Info.Id)
			if err != nil {
				txn.Rollback()
				return micro.NewException(structs2.DatabaseError, err)
			}
		} else {
			err = utils2.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, hierarchy.Info.Name,
				hierarchy.Info.Description, hierarchy.Info.Active, user, hierarchy.Info.Alias, "HIERARCHY")
			if err != nil {
				txn.Rollback()
				return micro.NewException(structs2.DatabaseError, err)
			}
			hierarchy.Info.Id = id
		}
		exception := action.saveReferences(hierarchy, txn, user)
		if exception != nil {
			txn.Rollback()
			return exception
		}
		action.savedHierarchies[idx] = hierarchy
	}

	err = txn.Commit()

	return micro.NewException(structs2.DatabaseError, err)
}

func (action *SaveHierarchiesAction) saveReferences(hierarchy structs.Hierarchy, tx *sql.Tx, user string) *micro.Exception {
	insertSqlRef := "INSERT INTO ref_hierarchies_types (hierarchy_id, index, object_type, action_by) " +
		"VALUES ($1, $2, $3, $4) "
	for _, entry := range hierarchy.Entries {
		err := utils2.ExecuteQueryWithTransaction(tx, insertSqlRef, hierarchy.Info.Id,
			entry.Index, entry.ObjectType, user)
		if err != nil {
			tx.Rollback()
			return micro.NewException(structs2.DatabaseError, err)
		}
	}

	return nil
}
