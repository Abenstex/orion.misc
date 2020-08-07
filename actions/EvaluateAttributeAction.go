package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"laniakea/dataStructures"
	"laniakea/micro"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	structs2 "orion.commons/structs"
	"orion.commons/utils"
	"orion.misc/structs"
	"time"
)

type EvaluateAttributeAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action EvaluateAttributeAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.EvaluateAttributeRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action EvaluateAttributeAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action EvaluateAttributeAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action EvaluateAttributeAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action EvaluateAttributeAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *EvaluateAttributeAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *EvaluateAttributeAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action EvaluateAttributeAction) SendEvents(request micro.IRequest) {

}

func (action EvaluateAttributeAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attribute/evaluate"
	var error = "orion/server/misc/error/attribute/evaluate"
	var requestSample = dataStructures.StructToJsonString(structs.EvaluateAttributeRequest{})
	var replySample = dataStructures.StructToJsonString(structs.EvaluateAttributeReply{})
	info := micro.ActionInformation{
		Name:           "EvaluateAttributeAction",
		Description:    "Evaluates attributes for an object using possible hierarchies if setup",
		RequestPath:    "orion/server/misc/request/attribute/evaluate",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.baseAction.ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   true,
	}

	return info
}

func (action *EvaluateAttributeAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *EvaluateAttributeAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	evaluateRequest := structs.EvaluateAttributeRequest{}
	json.Unmarshal(request, &evaluateRequest)

	return nil, nil
}

func (action *EvaluateAttributeAction) evaluateAttribute(request structs.GetAttributeDefinitionsRequest) (string, *micro.Exception) {
	/*query := "SELECT a.datatype, a.overwriteable, a.query, b.object_type, a.default_value," +
	" b.attr_value, c.hierarchy_id" +
	" FROM attributes a LEFT OUTER JOIN ref_attributes_objects b ON a.id = b.attr_id" +
	" LEFT OUTER JOIN ref_hierarchies_types c ON b.object_type=c.object_type" +
	" WHERE a.id=$1 and b.object_id=$2"*/

	return "", nil
}
