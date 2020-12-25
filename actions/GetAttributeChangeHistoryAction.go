package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/orion.commons/app"
	"github.com/abenstex/orion.commons/couchdb"
	http2 "github.com/abenstex/orion.commons/http"
	structs2 "github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"orion.misc/structs"
	"strconv"
	"time"
)

type FindAttributeChangeResponse struct {
	Docs []couchdb.HistoricizedAttributeDataChange `json:"docs"`
}

type GetAttributeChangeHistoryAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
}

func (action GetAttributeChangeHistoryAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs.GetAttributeValueChangeHistoryRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs2.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs2.RequestHeaderInvalid, err)
}

func (action GetAttributeChangeHistoryAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action GetAttributeChangeHistoryAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action GetAttributeChangeHistoryAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action GetAttributeChangeHistoryAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *GetAttributeChangeHistoryAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *GetAttributeChangeHistoryAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action GetAttributeChangeHistoryAction) SendEvents(request micro.IRequest) {

}

func (action GetAttributeChangeHistoryAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributehistory/get"
	var error = "orion/server/misc/error/attributehistory/get"
	var requestSample = dataStructures.StructToJsonString(structs.GetAttributeValueChangeHistoryRequest{})
	var replySample = dataStructures.StructToJsonString(structs.GetAttributeValueChangeHistoryReply{})

	info := micro.ActionInformation{
		Name:           "GetAttributeChangeHistoryAction",
		Description:    "GetAttributeChangeHistoryAction is used to get historicized attribute changes from CouchDB",
		RequestPath:    "orion/server/misc/request/attributehistory/get",
		ReplyPath:      dataStructures.JsonNullString{NullString: sql.NullString{String: reply, Valid: true}},
		ErrorReplyPath: dataStructures.JsonNullString{NullString: sql.NullString{String: error, Valid: true}},
		Version:        1,
		ClientId:       dataStructures.JsonNullString{NullString: sql.NullString{String: action.GetBaseAction().ID.String(), Valid: true}},
		HttpMethods:    []string{http.MethodPost, "OPTIONS"},
		RequestSample:  dataStructures.JsonNullString{NullString: sql.NullString{String: requestSample, Valid: true}},
		ReplySample:    dataStructures.JsonNullString{NullString: sql.NullString{String: replySample, Valid: true}},
		IsScriptable:   false,
	}

	return info
}

func (action *GetAttributeChangeHistoryAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *GetAttributeChangeHistoryAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	//fmt.Printf("History request: %v\n", string(request))

	dummy := structs.GetAttributeValueChangeHistoryRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyPath.String), &dummy
	}

	requests, err := action.readHistory(dummy)
	if err != nil {
		return structs2.NewErrorReplyHeaderWithOrionErr(structs2.NewOrionError(structs2.DatabaseError, err),
			action.ProvideInformation().ErrorReplyPath.String), &dummy
	}

	reply := structs.GetAttributeValueChangeHistoryReply{}
	reply.ChangedAttributes = requests
	reply.Header.Success = true

	return reply, &dummy
}

func (action GetAttributeChangeHistoryAction) readHistory(request structs.GetAttributeValueChangeHistoryRequest) ([]couchdb.HistoricizedAttributeDataChange, error) {
	config := couchdb.ReadCouchDbConfig()
	tmpDb := viper.GetString("couchdb.runtimeDatabase")
	config.Database = &tmpDb

	url := "http://" + *config.Host + ":" + strconv.Itoa(*config.Port) + "/" + *config.Database + "/_find"

	query := fmt.Sprintf("{\n   \"selector\": {\n"+
		"\"objectId\": {\n"+
		"\"$eq\": %v\n"+
		"}\n"+
		"},\n"+
		"\"fields\": [\"type\", \"referencedType\", \"receivedTime\", \"requestPath\", "+
		"\"hostAddress\", \"oldValue\", \"newValue\", \"attributeId\", \"objectId\"]\n}", request.ObjectId)

	//fmt.Printf("Query: %v\n", string(query))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}
	authString := base64.StdEncoding.EncodeToString([]byte(*config.User + ":" + *config.Password))
	numOfBytes := len(query)
	// "fields": ["_id", "_rev", "year", "title"],
	req.Header.Set("Content-Length", strconv.Itoa(numOfBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("withCredentials", "true")
	req.Header.Set("Authorization", "Basic "+authString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := FindAttributeChangeResponse{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Response: "+string(bodyBytes))
	//fmt.Printf("Requests size: "+strconv.Itoa(len(result.Requests)))

	return result.Docs, nil
}
