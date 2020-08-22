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
	laniakea "laniakea/utils"
	"net/http"
	"orion.commons/app"
	http2 "orion.commons/http"
	"orion.commons/structs"
	"orion.commons/utils"
	structs2 "orion.misc/structs"
	"time"
)

type SaveCategoriesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs2.Category
	saveRequest  structs2.SaveCategoriesRequest
}

func (action SaveCategoriesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.SaveCategoriesRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	action.saveRequest = dummy

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action SaveCategoriesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action SaveCategoriesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action SaveCategoriesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action SaveCategoriesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *SaveCategoriesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *SaveCategoriesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action SaveCategoriesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.SaveCategoriesRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("SaveCategoriesAction",
			action.GetBaseAction().Environment,
			true).Warn("RequestFailedEvent will be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		blerghEvent.Send(action.ProvideInformation().ErrorReplyPath.String, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	ids := make([]int64, 0, len(saveRequest.UpdatedCategories))
	for _, category := range saveRequest.UpdatedCategories {
		ids = append(ids, category.Info.Id)
	}
	event := structs2.CategorySavedEvent{
		Header:     *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		Categories: action.savedObjects,
		ObjectType: "CATEGORY",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("SaveCategoriesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic.String, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action SaveCategoriesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/category/save"
	var error = "orion/server/misc/error/category/save"
	var event = "orion/server/misc/event/category/save"
	var requestSample = dataStructures.StructToJsonString(structs2.SaveCategoriesRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	var eventSample = dataStructures.StructToJsonString(structs2.CategorySavedEvent{})
	info := micro.ActionInformation{
		Name:           "SaveCategoriesAction",
		Description:    "Saves categories and all necessary references to the database",
		RequestPath:    "orion/server/misc/request/category/save",
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

func (action *SaveCategoriesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *SaveCategoriesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)

	exception := action.saveObjects(action.saveRequest.UpdatedCategories, action.saveRequest.Header.User)
	if exception != nil {
		//fmt.Printf("Save Users error: %v\n", err)
		logging.GetLogger("SaveCategoriesAction",
			action.GetBaseAction().Environment,
			true).WithField("exception:", exception).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithException(exception,
			action.ProvideInformation().ErrorReplyPath.String), &action.saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyPath.String)
	reply.Success = true

	return reply, &action.saveRequest
}

func (action *SaveCategoriesAction) saveObjects(categories []structs2.Category, user string) *micro.Exception {
	insertSql := "INSERT INTO categories (name, description, active, action_by, " +
		"pretty_id, referenced_type, object_type) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id"
	updateSql := "UPDATE categories SET  name = $1, description = $2, active = $3, action_by = $4, " +
		"pretty_id = $5, referenced_type = $6, action_date=current_timestamp WHERE id = $7 "
	action.savedObjects = make([]structs2.Category, len(categories), len(categories))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return micro.NewException(structs.DatabaseError, err)
	}
	for idx, category := range categories {
		var id int64
		if category.Info.Id <= 0 {
			err = laniakea.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, category.Info.Name,
				category.Info.Description, category.Info.Active, user,
				category.Info.Alias, category.ReferencedType, "CATEGORY")
			if err != nil {
				logging.GetLogger("SaveCategoriesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not insert category")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
			category.Info.Id = id
		} else {
			err := laniakea.ExecuteQueryWithTransaction(txn, updateSql, category.Info.Name,
				category.Info.Description, category.Info.Active, user, category.Info.Alias,
				category.ReferencedType, category.Info.Id)
			if err != nil {
				logging.GetLogger("SaveCategoriesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not update category")
				txn.Rollback()
				return micro.NewException(structs.DatabaseError, err)
			}
		}

		action.savedObjects[idx] = category
	}
	err = txn.Commit()
	if err != nil {
		logging.GetLogger("SaveCategoriesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Commit failed")
		return micro.NewException(structs.DatabaseError, err)
	}

	return nil
}
