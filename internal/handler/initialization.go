package handler

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/sirupsen/logrus"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	if len(params.WorkspaceFolders) == 0 {
		return errors.New("length WorkspaceFolders is 0")
	}

	workspaceURI, err := uri.Parse(params.WorkspaceFolders[0].URI)
	if err != nil {
		logger.Error("Error parsing workspace URI", err)
		return err
	}
	h.yamllsConnector.CallInitialize(workspaceURI)

	h.chartStore = charts.NewChartStore(workspaceURI, charts.NewChart)

	return reply(ctx, lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindIncremental,
				OpenClose: true,
				Save: &lsp.SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", "$."},
				ResolveProvider:   false,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
	}, nil)
}

func (h *langHandler) initializationWithConfig() {
	configureLogLevel(h.helmlsConfig)
	h.chartStore.SetValuesFilesConfig(h.helmlsConfig.ValuesFilesConfig)
	configureYamlls(h)
}

func configureYamlls(h *langHandler) {
	config := h.helmlsConfig
	if config.YamllsConfiguration.Enabled {
		h.yamllsConnector = yamlls.NewConnector(config.YamllsConfiguration, h.connPool, h.documents)
		h.yamllsConnector.CallInitialize(h.chartStore.RootURI)
		h.yamllsConnector.InitiallySyncOpenDocuments(h.documents.GetAllDocs())
	}
}

func configureLogLevel(helmlsConfig util.HelmlsConfiguration) {
	if level, err := logrus.ParseLevel(helmlsConfig.LogLevel); err == nil {
		logger.SetLevel(level)
	} else {
		logger.Println("Error parsing log level", err)
	}
	if os.Getenv("LOG_LEVEL") == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}
}
