package actions

import (
	"database/sql"
	"encoding/json"
	"github.com/spf13/viper"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	"laniakea/mqtt"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	"orion.commons/structs"
	"orion.commons/utils"
	"time"
)

type DeleteStateAction struct {
	baseAction    micro.BaseAction
	MetricsStore  *utils.MetricsStore
	deleteRequest structs.DeleteRequest
	objectName    string
}

func (action *DeleteStateAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action DeleteStateAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DeleteStateAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DeleteStateAction) SendEvents(request micro.IRequest) {
	delRequest := request.(*structs.DeleteRequest)
	if !delRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DeleteStateAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(delRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs.DeletedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), request.GetHeader().SenderId, ""),
		ObjectId:   delRequest.ObjectId,
		ObjectType: "STATE",
		ObjectName: action.objectName,
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DeleteStateAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")), utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DeleteStateAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/state/delete"
	var error = "orion/server/misc/error/state/delete"
	var event = "orion/server/misc/event/state/delete"
	sampleRequest, sampleReply := action.sampleRequestReply()
	requestJson, _ := sampleRequest.ToString()
	replyJson, _ := sampleReply.MarshalJSON()
	info := micro.ActionInformation{
		Name:           "DeleteStateAction",
		Description:    "Delete a state from the database",
		RequestPath:    "orion/server/misc/request/state/delete",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestJson, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replyJson, Valid: true}},
		EventTopic:     dataStructures.JsonNullString{NullString: sql.NullString{String: event, Valid: true}},
	}

	return info
}

func (action DeleteStateAction) sampleRequestReply() (micro.IRequest, micro.IReply) {
	request := structs.DeleteRequest{
		Header:   micro.SampleRequestHeader(),
		ObjectId: 42,
	}
	reply := structs.NewReplyHeader("sample/topic")

	return &request, reply
}

func (action *DeleteStateAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DeleteStateAction) HeyHo(request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	env := action.GetBaseAction().Environment
	err := json.Unmarshal(request, &action.deleteRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}
	err = app.DefaultHandleActionRequest(request, &action.deleteRequest.Header, action, true)
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	err = action.getNameBeforeDelete(action.deleteRequest.ObjectId)
	if err != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}
	err = utils.DeleteObjectById(env, "states", action.deleteRequest.ObjectId, "DeleteStateAction")
	if err != nil {
		return structs.NewErrorReplyHeaderWithErr(err,
			action.ProvideInformation().ErrorReplyPath.String), &action.deleteRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.deleteRequest
}

func (action *DeleteStateAction) getNameBeforeDelete(objectId int64) error {
	query := "SELECT name FROM states WHERE id=$1"
	var name string
	row := action.GetBaseAction().Environment.Database.QueryRow(query, objectId)
	err := row.Scan(&name)

	action.objectName = name

	return err
}
