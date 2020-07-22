package actions

import (
	"context"
	"database/sql"
	"fmt"
	structs "laniakea/cache"
	"laniakea/dataStructures"
	"laniakea/logging"
	"laniakea/micro"
	"laniakea/utils"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	app2 "orion.commons/app"
	"orion.commons/couchdb"
	"orion.commons/http"
)

const ApplicationName = "ORION.Misc"
const ApplicationVersion = "0.0.1"
const HeartbeatTopic = "orion/server/heartbeat/misc"

type MiscApp struct {
	CacheManager structs.CacheManager
	AppInfo      micro.MicroServiceApplicationInformation
	Environment  utils.Environment
	Started      bool
	topicActions map[string]micro.Action
	timer        *time.Timer
	Token        *string
}

func (app *MiscApp) Init(configPath string) (utils.Environment, error) {
	environment, err := utils.NewEnvironment(configPath)
	if err != nil {
		return environment, err
	}

	environment.ApplicationId = viper.GetInt("general.applicationId")
	environment.ApplicationVersion = ApplicationVersion
	environment.ApplicationName = ApplicationName

	hostName, _ := os.Hostname()
	var startTime = utils.GetCurrentTimeStamp()
	var topic = viper.GetString("messagebus.baseTopic")
	var errorTopic = viper.GetString("messagebus.baseErrorTopic")
	var info []micro.ActionInformation

	app.AppInfo = micro.MicroServiceApplicationInformation{
		HostAddress:       utils.GetLocalIP(),
		HostName:          dataStructures.JsonNullString{NullString: sql.NullString{hostName, true}},
		AppName:           ApplicationName,
		AppVersion:        ApplicationVersion,
		AppInstance:       viper.GetInt("general.applicationId"),
		StartTime:         startTime,
		Description:       dataStructures.JsonNullString{NullString: sql.NullString{"ORION Misc Server Module", true}},
		BaseBusTopic:      dataStructures.JsonNullString{NullString: sql.NullString{String: topic, Valid: true}},
		BaseErrorTopic:    dataStructures.JsonNullString{NullString: sql.NullString{String: errorTopic, Valid: true}},
		ActionInformation: info,
		Company:           dataStructures.JsonNullString{NullString: sql.NullString{String: "Blackhole Software", Valid: true}},
		Port:              viper.GetInt("http.port"),
	}

	app.Environment = environment
	app.topicActions = make(map[string]micro.Action)

	return environment, nil
}

func (app *MiscApp) historicize(action micro.Action, receivedTime int64, requestPayload string, requestError error) {
	if viper.GetBool("history." + action.ProvideInformation().Name) {
		go func() {
			success := true
			if requestError != nil {
				success = false
			}
			err := couchdb.HistoricizeRequestReplyFromString(requestPayload, success, requestError,
				action.ProvideInformation().RequestPath, "BUS", receivedTime)
			if err != nil {
				logger := logging.GetLogger(ApplicationName, app.Environment, true)
				logger.WithError(err).Error("Historicizing request to CouchDB failed")
			}
		}()
	}
}

func (app *MiscApp) OnMessageReceived(client MQTT.Client, message MQTT.Message) {
	_, ok := app.topicActions[message.Topic()]
	receivedTime := utils.GetCurrentTimeStamp()
	if ok {
		action := app.topicActions[message.Topic()]

		ctx := context.TODO()
		go action.BeforeActionAsync(ctx, message.Payload())
		exception := action.BeforeAction(ctx, message.Payload())
		success := true
		requestPayload := string(message.Payload())
		var requestError error
		if exception != nil {
			requestError = fmt.Errorf(exception.ErrorText)
			success = false

			client.Publish(action.ProvideInformation().ErrorReplyPath.String, 0, false, exception)
		} else {
			iReply, iRequest := action.HeyHo(ctx, message.Payload())
			iRequest.HandleResult(iReply)
			header := iRequest.GetHeader()
			header.UpdateReceivedTime(receivedTime)
			iRequest.UpdateHeader(header)
			//topic, reply, ok := functionMap[message.Topic()].HeyHo(message.Payload())
			jsonWurst, err := iReply.MarshalJSON()

			if err == nil {
				client.Publish(action.ProvideInformation().ReplyPath.String, 0, false, jsonWurst)
			} else {
				client.Publish(action.ProvideInformation().ErrorReplyPath.String, 0, false, err.Error())
			}
			success = iReply.Successful()
			requestPayload, _ = iRequest.ToString()
			if success {
				exception = action.AfterAction(ctx, &iReply, &iRequest)
				if exception != nil {
					client.Publish(action.ProvideInformation().ErrorReplyPath.String, 0, false, exception)
				}
				go action.AfterActionAsync(ctx, iReply, iRequest)
			} else {
				requestError = fmt.Errorf(iReply.Error())
			}
			go action.SendEvents(iRequest)
		}
		//err := app2.DefaultHandleAction(action, message.Payload(), client)
		app.historicize(action, receivedTime, requestPayload, requestError)
	} else {
		errorReply := fmt.Sprintf("No handler was found for topic %s on "+
			"application %s version %s running on %s",
			message.Topic(), app.AppInfo.AppName, app.AppInfo.AppVersion,
			app.AppInfo.HostAddress)
		fmt.Println(errorReply)
		client.Publish(app.AppInfo.BaseErrorTopic.String, byte(viper.GetInt("messageBus.replyQos")), false, errorReply)
	}
}

func (app *MiscApp) ProvideApplicationInformation() micro.MicroServiceApplicationInformation {
	return app.AppInfo
}

func (app *MiscApp) StartApplication(actions []micro.Action) error {
	topicActions, err := app2.DefaultStartApplication(app, HeartbeatTopic, actions, &app.AppInfo.ActionInformation, app.Environment, app.OnMessageReceived)

	if err != nil {
		logger := logging.GetLogger(ApplicationName, app.Environment, true)
		logger.WithError(err).Fatal("Could not start application! Exiting...")

		os.Exit(666)
	}

	app.topicActions = topicActions
	//app.Started = true

	logging.GetLogger(ApplicationName, app.Environment, true).Info("Server started and is ready for requests with PID " + strconv.Itoa(os.Getpid()))

	return nil
}

func (app *MiscApp) StopApplication() error {
	logger := logging.GetLogger(app.AppInfo.AppName, app.Environment, true)
	logger.Debug("Stopping " + ApplicationName + " " + ApplicationVersion)
	//utils.StopCommunication()

	return app.Environment.Database.Close()
}

func (app *MiscApp) RegisterApplication() error {
	url := viper.GetString("http.registrationURL")
	request := micro.NewRegisterMicroServiceRequest(app.AppInfo)
	jwt, err := http.RegisterApp(url, request)
	if err != nil {
		return err
	}

	//fmt.Printf("Reply: %v\n", &jwt)
	app.Token = jwt
	for _, value := range app.topicActions {
		dummyBaseAction := value.GetBaseAction()
		dummyBaseAction.Token = jwt
		value.InitBaseAction(dummyBaseAction)
	}

	app.Started = true

	return nil
}

func (app *MiscApp) UnregisterApplication() error {
	request := micro.NewUnregisterMicroServiceRequest(app.AppInfo, *app.Token)
	url := viper.GetString("http.unregistrationURL")
	_, err := http.UnregisterApp(url, request)

	return err
}
