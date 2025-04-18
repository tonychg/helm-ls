package handler

import (
	"context"
	"encoding/json"

	// "reflect"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) (err error) {
	logger.Println("Reload yaml configuration:", params)
	h.helmlsConfig.YamllsConfiguration.YamllsSettings = params.Settings
	h.configureYamlls(ctx)
	return nil
}

func (h *langHandler) retrieveWorkspaceConfiguration(ctx context.Context) {
	logger.Debug("Calling workspace/configuration")
	configurationParams := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{{Section: "helm-ls"}},
	}

	rawResult, err := h.client.Configuration(ctx, &configurationParams)
	if err != nil {
		logger.Println("Error calling workspace/configuration", err)
		h.initializationWithConfig(ctx)
		return
	}

	h.helmlsConfig = parseWorkspaceConfiguration(rawResult, h.helmlsConfig)
	logger.Println("Workspace configuration:", h.helmlsConfig)
	h.initializationWithConfig(ctx)
}

func parseWorkspaceConfiguration(rawResult []interface{}, currentConfig util.HelmlsConfiguration) (result util.HelmlsConfiguration) {
	logger.Debug("Raw Workspace configuration:", rawResult)

	if len(rawResult) == 0 {
		logger.Println("Workspace configuration is empty")
		return currentConfig
	}

	jsonResult, err := json.Marshal(rawResult[0])
	if err != nil {
		logger.Println("Error marshalling workspace/configuration", err)
		return currentConfig
	}

	result = currentConfig
	err = json.Unmarshal(jsonResult, &result)
	if err != nil {
		logger.Println("Error unmarshalling workspace/configuration", err)
		return currentConfig
	}
	return result
}
