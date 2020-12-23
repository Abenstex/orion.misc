package main

import (
	"flag"
	"fmt"
	"laniakea/logging"
	"laniakea/micro"
	laniakeautils "laniakea/utils"
	"syscall"
	"time"

	"github.com/ztrue/shutdown"
	"orion.commons/utils"
	"orion.misc/actions"
)

func main() {
	configPath := flag.String("config", "", "The absolute path to the config file")
	flag.Parse()

	app := actions.MiscApp{}
	app.AppInfo.ActionInformation = make([]micro.ActionInformation, 25)

	shutdown.Add(func() {
		_ = app.UnregisterApplication()
		time.Sleep(5 * time.Second)
		_ = app.StopApplication()
	})

	env, err := app.Init(*configPath)
	if err != nil {
		logging.GetLogger("general_errors", env, true).WithError(err).Fatal("Could not init application server")
		_ = app.StopApplication()
	}
	metricsStore := new(utils.MetricsStore)

	baseAction := micro.BaseAction{
		Environment: app.Environment,
		ID:          laniakeautils.NewUuid(),
		Request:     nil,
		Token:       nil,
	}

	saveStatesAction := actions.SaveStatesAction{MetricsStore: metricsStore}
	saveStatesAction.InitBaseAction(baseAction)
	deleteStateAction := actions.DeleteStateAction{MetricsStore: metricsStore}
	deleteStateAction.InitBaseAction(baseAction)
	getStatesAction := actions.GetStatesAction{MetricsStore: metricsStore}
	getStatesAction.InitBaseAction(baseAction)
	defineAttributesAction := actions.DefineAttributesAction{MetricsStore: metricsStore}
	defineAttributesAction.InitBaseAction(baseAction)
	deleteAttributeDefinitionAction := actions.DeleteAttributeDefinitionAction{MetricsStore: metricsStore}
	deleteAttributeDefinitionAction.InitBaseAction(baseAction)
	getAttributeDefinitionsAction := actions.GetAttributeDefinitionsAction{MetricsStore: metricsStore}
	getAttributeDefinitionsAction.InitBaseAction(baseAction)
	setAttributeValueAction := actions.SetAttributeValueAction{MetricsStore: metricsStore}
	setAttributeValueAction.InitBaseAction(baseAction)
	deleteAttributeValueAction := actions.DeleteAttributeValueAction{MetricsStore: metricsStore}
	deleteAttributeValueAction.InitBaseAction(baseAction)
	getAttributeValuesAction := actions.GetAttributeValuesPerObjectAction{MetricsStore: metricsStore}
	getAttributeValuesAction.InitBaseAction(baseAction)
	getAttributeChangeHistoryAction := actions.GetAttributeChangeHistoryAction{MetricsStore: metricsStore}
	getAttributeChangeHistoryAction.InitBaseAction(baseAction)
	saveHierarchiesAction := actions.SaveHierarchiesAction{MetricsStore: metricsStore}
	saveHierarchiesAction.InitBaseAction(baseAction)
	deleteHierarchyAction := actions.DeleteHierarchyAction{MetricsStore: metricsStore}
	deleteHierarchyAction.InitBaseAction(baseAction)
	getHierarchiesAction := actions.GetHierarchiesAction{MetricsStore: metricsStore}
	getHierarchiesAction.InitBaseAction(baseAction)
	saveParametersAction := actions.SaveParametersAction{MetricsStore: metricsStore}
	saveParametersAction.InitBaseAction(baseAction)
	deleteParameterAction := actions.DeleteParameterAction{MetricsStore: metricsStore}
	deleteParameterAction.InitBaseAction(baseAction)
	getParametersAction := actions.GetParametersAction{MetricsStore: metricsStore}
	getParametersAction.InitBaseAction(baseAction)
	saveCategoriesAction := actions.SaveCategoriesAction{MetricsStore: metricsStore}
	saveCategoriesAction.InitBaseAction(baseAction)
	getCategoriesAction := actions.GetCategoriesAction{MetricsStore: metricsStore}
	getCategoriesAction.InitBaseAction(baseAction)
	deleteCategoryAction := actions.DeleteCategoryAction{MetricsStore: metricsStore}
	deleteCategoryAction.InitBaseAction(baseAction)
	getObjectsPerCategoryAction := actions.GetObjectsPerCategoryAction{MetricsStore: metricsStore}
	getObjectsPerCategoryAction.InitBaseAction(baseAction)
	saveCategoriesReferencesAction := actions.SaveObjectCategoryReferenceAction{MetricsStore: metricsStore}
	saveCategoriesReferencesAction.InitBaseAction(baseAction)
	saveObjectTypeCustomizationsAction := actions.SaveObjectTypeCustomizationAction{MetricsStore: metricsStore}
	saveObjectTypeCustomizationsAction.InitBaseAction(baseAction)
	getObjectTypeCustomizationsAction := actions.GetObjectTypeCustomizationsAction{MetricsStore: metricsStore}
	getObjectTypeCustomizationsAction.InitBaseAction(baseAction)
	saveStateTransitionRulesAction := actions.SaveStateTransitionRulesAction{MetricsStore: metricsStore}
	saveStateTransitionRulesAction.InitBaseAction(baseAction)
	getStateTransitionRulesAction := actions.GetStateTransitionRulesAction{MetricsStore: metricsStore}
	getStateTransitionRulesAction.InitBaseAction(baseAction)

	services := []micro.Action{&saveStatesAction, &deleteStateAction, &getStatesAction, &defineAttributesAction,
		&deleteAttributeDefinitionAction, &getAttributeDefinitionsAction, &setAttributeValueAction, &deleteAttributeValueAction,
		&getAttributeValuesAction, &getAttributeChangeHistoryAction, &saveHierarchiesAction,
		&deleteHierarchyAction, &getHierarchiesAction, &saveParametersAction, &deleteParameterAction,
		&getParametersAction, &saveCategoriesAction, &getCategoriesAction, &deleteCategoryAction,
		&saveCategoriesReferencesAction, &getObjectsPerCategoryAction, &saveObjectTypeCustomizationsAction,
		&getObjectTypeCustomizationsAction, &saveStateTransitionRulesAction, &getStateTransitionRulesAction}

	_ = app.StartApplication(services)
	err = app.RegisterApplication()
	if err != nil {
		fmt.Printf("An error occurred during registration: %v\nShutting down...\n", err)
		logging.GetLogger("general_errors", env, true).WithError(err).Fatal("Could not register application. Shutting down...")
		_ = app.StopApplication()
	}

	shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)
}
