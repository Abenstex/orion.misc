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

type SaveParametersAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.Parameter
}

func (action SaveParametersAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveParametersRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action SaveParametersAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveParametersAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveParametersAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveParametersAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveParametersAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveParametersAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveParametersAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveParametersRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveParametersAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.Parameters))
	for _, parameter := range saveRequest.Parameters {
		ids = append(ids, parameter.Info.Id)
	}
	event := structs2.ParameterSavedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Parameters: action.savedObjects,
		ObjectType: "PARAMETER",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveParametersAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveParametersAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/parameter/save"
	var error = "orion/server/misc/error/parameter/save"
	var event = "orion/server/misc/event/parameter/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveParametersRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.ParameterSavedEvent{})
	info := micro.ActionInformation{
		Name:           "SaveParametersAction",
		Description:    "Saves PARAMETER and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/parameter/save",
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

func (action *SaveParametersAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveParametersAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	saveRequest := structs2.SaveParametersRequest{}

	err := json.Unmarshal(request, &saveRequest)
	//fmt.Printf("Saverequest: %v\n", string(request))
	if err != nil {
		return structs.NewErrorReplyHeaderWithException(micro.NewException(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	exception := action.saveObjects(saveRequest.Parameters, saveRequest.Header.User)
	if exception != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SaveParametersAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithException(exception,
			action.ProvideInformation().ErrorReplyPath.String), &saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &saveRequest
}

func (action *SaveParametersAction) saveObjects(parameters []structs2.Parameter, user string) *micro.Exception {
	insertSql := "INSERT INTO parameters (name, description, active, action_by, " +
		"pretty_id, value, object_type) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id"
	updateSql := "UPDATE parameters SET  name = $1, description = $2, active = $3, action_by = $4, " +
		"pretty_id = $5, value = $6 WHERE id = $7 "
	action.savedObjects = make([]structs2.Parameter, len(parameters), len(parameters))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return micro.NewException(structs.DatabaseError, err)
	}
	for idx, parameter := range parameters {
		var id int64
		if parameter.Info.Id <= 0 {
			err = laniakea.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, parameter.Info.Name,
				parameter.Info.Description, parameter.Info.Active, user,
				parameter.Info.Alias, parameter.Value, "PARAMETER")
			if err != nil {
				logging.GetLogger("SaveParametersAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not insert parameter")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
			parameter.Info.Id = id
		} else {
			err := laniakea.ExecuteQueryWithTransaction(txn, updateSql, parameter.Info.Name,
				parameter.Info.Description, parameter.Info.Active, user, parameter.Info.Alias,
				parameter.Value, parameter.Info.Id)
			if err != nil {
				logging.GetLogger("SaveParametersAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not update parameter")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
		}

		action.savedObjects[idx] = parameter
	}
	err = txn.Commit()
	if err != nil {
		return micro.NewException(structs.DatabaseError, err)
	}

	return nil
}
