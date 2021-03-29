package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/lib/pq"
	"net/http"
	"orion.misc/structs"
	"time"
)

type EvaluateAttributeAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

type AttributeToEvaluate struct {
	DataType      string
	Overwriteable bool
	Query         *string
	ObjectType    string
	DefaultValue  *string
	ListOfValues  *[]string
	Value         *string
	HierarchyId   *int64
}

func (action EvaluateAttributeAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.EvaluateAttributeRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	if dummy.AttributeId <= 0 || dummy.ObjectId <= 0 {
		return micro.NewException(structs2.MissingParameterError, fmt.Errorf("not all parameters (attribute_id and object_id) were provided"))
	}

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
		Name:            "EvaluateAttributeAction",
		Description:     "Evaluates attributes for an object using possible hierarchies if setup",
		RequestTopic:    "orion/server/misc/request/attribute/evaluate",
		ReplyTopic:      reply,
		ErrorReplyTopic: error,
		Version:         1,
		ClientId:        action.baseAction.ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		IsScriptable:    true,
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

func (action *EvaluateAttributeAction) evaluateAttribute(request structs.EvaluateAttributeRequest) (string, *micro.Exception) {
	query := "SELECT a.datatype, a.overwriteable, a.query, b.object_type, a.default_value, a.list_of_values," +
		" b.attr_value, c.hierarchy_id" +
		" FROM attributes a LEFT OUTER JOIN ref_attributes_objects b ON a.id = b.attr_id" +
		" LEFT OUTER JOIN ref_hierarchies_types c ON b.object_type=c.object_type" +
		" WHERE a.id=$1 and b.object_id=$2"

	rows, err := action.GetBaseAction().Environment.Database.Query(query, request.AttributeId, request.ObjectId)
	if err != nil {
		return "", micro.NewException(structs2.DatabaseError, err)
	}
	defer rows.Close()

	attribute := AttributeToEvaluate{}
	err = rows.Scan(&attribute.DataType, &attribute.Overwriteable, &attribute.Query, &attribute.ObjectType,
		&attribute.DefaultValue, pq.Array(&attribute.ListOfValues), &attribute.Value, &attribute.HierarchyId)

	if !attribute.Overwriteable && attribute.Value != nil {
		return *attribute.Value, nil
	}

	return "", nil
}
