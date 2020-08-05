package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
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

type SaveStatesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedStates  []structs.State
}

func (action SaveStatesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveStatesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Errorf("error unmarshalling the request: %v\n", string(request))
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action SaveStatesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveStatesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveStatesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveStatesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveStatesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveStatesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveStatesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveStatesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveStatesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.UpdatedStates))
	for _, state := range saveRequest.UpdatedStates {
		ids = append(ids, state.Info.Id)
	}
	event := structs.SavedStatesEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), request.GetHeader().SenderId, ""),
		States:     action.savedStates,
		ObjectType: "STATE",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveStatesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveStatesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/state/save"
	var error = "orion/server/misc/error/state/save"
	var event = "orion/server/misc/event/state/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "SaveStatesAction",
		Description:    "Saves states to the database",
		RequestPath:    "orion/server/misc/request/state/save",
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

func (action *SaveStatesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveStatesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs.SaveStatesRequest{}

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	err = action.saveStates(saveRequest.UpdatedStates, saveRequest.OriginalState, saveRequest.Header.User)
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name,
			action.GetBaseAction().Environment,
			true).WithError(err).Error("Data could not be saved")
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	reply := structs2.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveStatesAction) saveStates(updatedStates []structs.State, originalStates []structs.State, user string) error {
	insertSql := "INSERT INTO states (name, description, active, action_by, pretty_id, " +
		"referenced_type, object_available, substate, default_state, object_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id"
	updateSql := "UPDATE states SET name = $1, description = $2, active = $3, " +
		"pretty_id = $4, referenced_type = $5, object_available = $6, substate = $7, default_state = $8 WHERE id = $9"

	action.savedStates = make([]structs.State, len(updatedStates), len(updatedStates))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return err
	}

	for idx, state := range updatedStates {
		var id int64
		if state.Info.Id <= 0 {
			err = utils2.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, state.Info.Name, state.Info.Description,
				state.Info.Active, user, state.Info.Alias, state.ReferencedType, state.ObjectAvailable, state.Substate,
				state.DefaultState, "STATE")
			if err != nil {
				txn.Rollback()
				return err
			}
			state.Info.Id = id
		} else {
			err := utils2.ExecuteQueryWithTransaction(txn, updateSql, state.Info.Name,
				state.Info.Description, state.Info.Active, state.Info.Alias,
				state.ReferencedType, state.ObjectAvailable, state.Substate,
				state.DefaultState, state.Info.Id)
			if err != nil {
				txn.Rollback()
				return err
			}
		}
		var origProfile = action.getOriginalState(state.Info.Id, originalStates)
		err = action.saveReferences(txn, state.Info.Id, state.Substates, origProfile.Substates, user)
		if err != nil {
			txn.Rollback()
			return err
		}
		action.savedStates[idx] = state
	}

	return txn.Commit()
}

func (action *SaveStatesAction) getOriginalState(objectId int64, originalStates []structs.State) structs.State {
	if objectId == -1 || originalStates == nil || len(originalStates) == 0 {
		return structs.State{}
	}
	for _, state := range originalStates {
		if objectId == state.Info.Id {
			return state
		}
	}

	return structs.State{}
}

func (action SaveStatesAction) saveReferences(txn *sql.Tx, stateId int64, updatedStates []int64, originalStates []int64, user string) error {
	compareResult := dataStructures.CompareInt64Slices(updatedStates, originalStates, true)
	// NotInSlice1 : IDs NICHT in updatedIDs aber in origIds -> delete
	// NotInSlice2 : IDs in updatedIDs aber NICHT in origIds -> insert
	deleteReferenceSql := "DELETE FROM ref_states_substates WHERE state_id=$1 AND substate_id = ANY($2::bigint[]) "

	if len(compareResult.NotInSlice1) > 0 {
		logging.GetLogger(action.ProvideInformation().Name, action.GetBaseAction().Environment, false).
			WithFields(logrus.Fields{"stateId": stateId}).Debug("Deleting updatedSubstates references for state")
		err := utils2.ExecuteQueryWithTransaction(txn, deleteReferenceSql, stateId, pq.Array(compareResult.NotInSlice1))
		if err != nil {
			return err
		}
	}
	columns := []string{"state_id", "substate_id", "action_by"}
	for _, refId := range compareResult.NotInSlice2 {
		err := utils2.ExecuteInsertWithTransaction(txn, "ref_states_substates", columns, stateId, refId, user)
		if err != nil {
			return err
		}
	}

	return nil
}
