package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
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

type SaveStateTransitionRulesAction struct {
	baseAction                micro.BaseAction
	MetricsStore              *utils.MetricsStore
	savedStateTransitionRules []structs.StateTransitionRule
}

func (action SaveStateTransitionRulesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.SaveStateTransitionRulesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action SaveStateTransitionRulesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveStateTransitionRulesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveStateTransitionRulesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveStateTransitionRulesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveStateTransitionRulesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveStateTransitionRulesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveStateTransitionRulesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs.SaveStateTransitionRulesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveStateTransitionRulesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs2.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.UpdatedStateTransitionRules))
	for _, state := range saveRequest.UpdatedStateTransitionRules {
		ids = append(ids, state.Info.Id)
	}
	event := structs.SavedStateTransitionRulesEvent{
		Header:               *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		StateTransitionRules: action.savedStateTransitionRules,
		ObjectType:           "STATE_TRANSITION_RULE",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveStateTransitionRulesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveStateTransitionRulesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/statetransitionrule/save"
	var error = "orion/server/misc/error/statetransitionrule/save"
	var event = "orion/server/misc/event/statetransitionrule/save"
	var requestSample = dataStructures.StructToJsonString(micro.RegisterMicroServiceRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:           "SaveStateTransitionRulesAction",
		Description:    "Saves statetransitionrules to the database",
		RequestPath:    "orion/server/misc/request/statetransitionrule/save",
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

func (action *SaveStateTransitionRulesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveStateTransitionRulesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs.SaveStateTransitionRulesRequest{}

	err := json.Unmarshal(request, &saveRequest)
	//fmt.Printf("Saverequest: %v\n", string(request))
	if err != nil {
		logging.GetLogger(action.ProvideInformation().Name, action.baseAction.Environment, true).WithError(err).Error("Could not unmarshal request")
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	err = action.saveStateTransitionRules(saveRequest.UpdatedStateTransitionRules, saveRequest.Header.User)
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

func (action *SaveStateTransitionRulesAction) saveStateTransitionRules(updatedRules []structs.StateTransitionRule, user string) error {
	updateSql := "UPDATE state_transition_rules SET name = $1, description = $2, active = $3, action_by = $4, " +
		"pretty_id = $5, from_state = $6, to_states = $7 WHERE id = $8 "

	action.savedStateTransitionRules = make([]structs.StateTransitionRule, len(updatedRules), len(updatedRules))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return err
	}

	insertColumns := []string{"name", "description", "active", "action_by", "pretty_id", "object_type", "from_state", "to_states"}

	for idx, rule := range updatedRules {
		if rule.Info.Id >= 0 {
			err := utils2.ExecuteQueryWithTransaction(txn, updateSql, rule.Info.Name,
				rule.Info.Description, rule.Info.Active, user, rule.Info.Alias, rule.FromState, pq.Array(rule.ToStates), rule.Info.Id)
			if err != nil {
				txn.Rollback()
				return err
			}
		} else {
			err = utils2.ExecuteInsertWithTransaction(txn, "state_transition_rules", insertColumns, rule.Info.Name, rule.Info.Description,
				rule.Info.Active, user, rule.Info.Alias, "STATE_TRANSITION_RULE", rule.FromState, pq.Array(rule.ToStates))
			if err != nil {
				txn.Rollback()
				return err
			}
		}
		action.savedStateTransitionRules[idx] = rule
	}

	return txn.Commit()

	return nil
}
