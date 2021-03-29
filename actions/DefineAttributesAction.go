package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/abenstex/laniakea/dataStructures"
	"github.com/abenstex/laniakea/logging"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/laniakea/mongodb"
	"github.com/abenstex/laniakea/mqtt"
	laniakea "github.com/abenstex/laniakea/utils"
	"github.com/abenstex/orion.commons/app"
	http2 "github.com/abenstex/orion.commons/http"
	"github.com/abenstex/orion.commons/structs"
	"github.com/abenstex/orion.commons/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	structs2 "orion.misc/structs"
	"time"
)

type DefineAttributesAction struct {
	baseAction   micro.BaseAction
	MetricsStore *utils.MetricsStore
	savedObjects []structs.AttributeDefinition
	startedTime  int64
}

func (action DefineAttributesAction) BeforeAction(ctx context.Context, request []byte) *micro.Exception {
	dummy := structs2.DefineAttributeRequest{}
	err := json.Unmarshal(request, &dummy)
	if err != nil {
		return micro.NewException(structs.UnmarshalError, err)
	}
	err = app.DefaultHandleActionRequest(request, &dummy.Header, &action, true)

	return micro.NewException(structs.RequestHeaderInvalid, err)
}

func (action DefineAttributesAction) BeforeActionAsync(ctx context.Context, request []byte) {

}

func (action DefineAttributesAction) AfterAction(ctx context.Context, reply *micro.IReply, request *micro.IRequest) *micro.Exception {
	return nil
}

func (action DefineAttributesAction) AfterActionAsync(ctx context.Context, reply micro.IReply, request micro.IRequest) {

}

func (action DefineAttributesAction) GetBaseAction() micro.BaseAction {
	return action.baseAction
}

func (action *DefineAttributesAction) SetHttpRequest(request *http.Request) {
	action.baseAction.Request = request
}

func (action *DefineAttributesAction) InitBaseAction(baseAction micro.BaseAction) {
	action.baseAction = baseAction
}

func (action DefineAttributesAction) SendEvents(request micro.IRequest) {
	saveRequest := request.(*structs2.DefineAttributeRequest)
	if !saveRequest.Header.WasExecutedSuccessfully {
		logging.GetLogger("DefineAttributesAction",
			action.GetBaseAction().Environment,
			true).Warn("Events won't be sent because the request was not successfully executed")
		blerghEvent := structs.NewRequestFailedEvent(saveRequest, action.ProvideInformation(), action.baseAction.ID.String(), "")
		_ = blerghEvent.Send(action.ProvideInformation().ErrorReplyTopic, byte(viper.GetInt("messageBus.publishEventQos")),
			utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
		return
	}
	event := structs2.AttributeDefinitionSavedEvent{
		Header:               *micro.NewEventHeaderForAction(action.ProvideInformation(), saveRequest.Header.SenderId, ""),
		AttributeDefinitions: action.savedObjects,
		ObjectType:           "AttributeDefinition",
	}

	json, err := event.ToJsonString()
	if err != nil {
		logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, true).WithError(err).Error("Could not send events")

		return
	}
	mqtt.Publish(action.ProvideInformation().EventTopic, json, byte(viper.GetInt("messageBus.publishEventQos")),
		utils.GetDefaultMqttConnectionOptionsWithIdPrefix(action.ProvideInformation().Name))
}

func (action DefineAttributesAction) ProvideInformation() micro.ActionInformation {
	var reply = "orion/server/misc/reply/attributedefinition/save"
	var errorSubject = "orion/server/misc/error/attributedefinition/save"
	var event = "orion/server/misc/event/attributedefinition/save"
	var requestSample = dataStructures.StructToJsonString(structs2.DefineAttributeRequest{})
	var replySample = dataStructures.StructToJsonString(micro.ReplyHeader{})
	info := micro.ActionInformation{
		Name:            "DefineAttributesAction",
		Description:     "Saves AttributeDefinition and all necessary references to the database",
		RequestTopic:    "orion/server/misc/request/attributedefinition/save",
		ReplyTopic:      reply,
		ErrorReplyTopic: errorSubject,
		Version:         1,
		ClientId:        action.GetBaseAction().ID.String(),
		HttpMethods:     []string{http.MethodPost, "OPTIONS"},
		EventTopic:      event,
		RequestSample:   &requestSample,
		ReplySample:     &replySample,
		IsScriptable:    false,
	}

	return info
}

func (action *DefineAttributesAction) HandleWebRequest(writer http.ResponseWriter, request *http.Request) {
	action.SetHttpRequest(request)
	http2.HandleHttpRequest(writer, request, action)
}

func (action *DefineAttributesAction) HeyHo(ctx context.Context, request []byte) (micro.IReply, micro.IRequest) {
	start := time.Now()
	defer action.MetricsStore.HandleActionMetric(start, action.GetBaseAction().Environment, action.ProvideInformation(), *action.baseAction.Token)
	action.startedTime = laniakea.GetCurrentTimeStamp()

	saveRequest := structs2.DefineAttributeRequest{}

	err := json.Unmarshal(request, &saveRequest)
	if err != nil {
		return structs.NewErrorReplyHeaderWithOrionErr(structs.NewOrionError(structs.UnmarshalError, err),
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	orionErr := action.saveObjects(ctx, saveRequest.UpdatedAttributeDefinitions, saveRequest.Header.Comment, saveRequest.Header.User)
	if orionErr != nil {
		logging.GetLogger("DefineAttributesAction",
			action.GetBaseAction().Environment,
			true).WithError(orionErr.Error).Error("Data could not be saved")
		return structs.NewErrorReplyHeaderWithOrionErr(orionErr,
			action.ProvideInformation().ErrorReplyTopic), &saveRequest
	}

	reply := structs.NewReplyHeader(action.ProvideInformation().ReplyTopic)
	reply.Success = true

	return reply, &saveRequest
}

func (action *DefineAttributesAction) archiveAndReplaceObject(ctx context.Context, object structs.AttributeDefinition) error {
	var objectToArchive structs.AttributeDefinition
	result, err := mongodb.ReplaceAndFindOneById(ctx, action.baseAction.Environment.MongoDbConnection, "attribute_definitions", object.ID.Hex(), object)
	if err != nil {
		return err
	}
	err = result.Decode(&objectToArchive)
	if err != nil {
		return err
	}
	objectToArchive.Info.ChangeDate = &action.startedTime
	_, err = mongodb.InsertOne(context.Background(), action.baseAction.Environment.MongoDbArchiveConnection, "attribute_definitions", objectToArchive)

	return err
}

func (action *DefineAttributesAction) saveObjects(ctx context.Context, updatedObjects []structs.AttributeDefinition, comment, user string) *structs.OrionError {
	newCtx := context.WithValue(ctx, "objects", updatedObjects)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		objects := sessCtx.Value("objects").([]structs.AttributeDefinition)
		for _, object := range objects {
			if object.Info.CreatedDate == 0 {
				object.Info.CreatedDate = laniakea.GetCurrentTimeStamp()
			}
			if object.ID == nil || object.ID.IsZero() {
				_, err := mongodb.InsertOne(sessCtx, action.baseAction.Environment.MongoDbConnection, "attribute_definitions", object)
				if err != nil {
					return nil, err
				}
			} else {
				object.Info.UserComment = &comment
				object.Info.User = &user
				object.Info.ChangeDate = &action.startedTime

				err := action.archiveAndReplaceObject(sessCtx, object)
				if err != nil {
					return nil, err
				}
			}
			action.savedObjects = append(action.savedObjects, object)
		}

		return nil, nil
	}
	_, err := mongodb.PerformQueriesInTransaction(newCtx, action.baseAction.Environment.MongoDbConnection, callback)
	if err != nil {
		return structs.NewOrionError(structs.DatabaseError, fmt.Errorf("error executing queries in transaction: %v", err))
	}

	return nil
}

/*func (action *DefineAttributesAction) saveObjects(updatedObjects []structs.AttributeDefinition, originalObjects []structs.AttributeDefinition, user string) error {
	insertSql := "INSERT INTO attributes (name, description, active, action_by, pretty_id, datatype, " +
		" overwriteable, allowed_object_types, list_of_values, numeric_from, numeric_to, " +
		" query, object_type, default_value, assign_during_object_creation) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id"
	updateSql := "UPDATE attributes SET name = $1, description = $2, active = $3, action_by = $4, " +
		"pretty_id = $5, overwriteable = $6, allowed_object_types = $7, list_of_values = $8, " +
		"numeric_from = $9, numeric_to = $10, query = $11, datatype = $12, default_value=$13, assign_during_object_creation = $14 WHERE id = $15 "
	action.savedObjects = make([]structs.AttributeDefinition, len(updatedObjects), len(updatedObjects))

	txn, err := action.GetBaseAction().Environment.Database.Begin()
	if err != nil {
		if txn != nil {
			txn.Rollback()
		}
		return err
	}
	for idx, updatedObject := range updatedObjects {
		var id int64
		if updatedObject.Info.Id <= 0 {
			err = laniakea.ExecuteInsertWithTransactionWithAutoId(txn, insertSql, &id, updatedObject.Info.Name,
				updatedObject.Info.Description, updatedObject.Info.Active, user,
				updatedObject.Info.Alias, updatedObject.DataType, updatedObject.Overwriteable, pq.Array(updatedObject.AllowedObjectTypes),
				pq.Array(updatedObject.ListOfValues), updatedObject.NumericFrom, updatedObject.NumericTo,
				updatedObject.Query, "ATTRIBUTE", updatedObject.DefaultValue, updatedObject.AssignDuringObjectCreation)
			if err != nil {
				logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not insert user")
				txn.Rollback()
				return err
			}
			updatedObject.Info.Id = id
		} else {
			err := laniakea.ExecuteQueryWithTransaction(txn, updateSql, updatedObject.Info.Name,
				updatedObject.Info.Description, updatedObject.Info.Active, user, updatedObject.Info.Alias, updatedObject.Overwriteable,
				pq.Array(updatedObject.AllowedObjectTypes), pq.Array(updatedObject.ListOfValues),
				updatedObject.NumericFrom, updatedObject.NumericTo, updatedObject.Query, updatedObject.DataType,
				updatedObject.DefaultValue, updatedObject.AssignDuringObjectCreation, updatedObject.Info.Id)
			if err != nil {
				logging.GetLogger("DefineAttributesAction", action.GetBaseAction().Environment, false).WithError(err).Error("Could not update user")
				txn.Rollback()
				return err
			}
		}

		action.savedObjects[idx] = updatedObject
	}

	return txn.Commit()
}*/
