//go:build integration

package yamlls

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

type jsonRpcDiagnostics struct {
	Params  lsp.PublishDiagnosticsParams `json:"params"`
	Jsonrpc string                       `json:"jsonrpc"`
	Method  string                       `json:"method"`
}

type readWriteCloseMock struct {
	diagnosticsChan chan lsp.PublishDiagnosticsParams
}

func (proc readWriteCloseMock) Read(p []byte) (int, error) {
	return 1, nil
}

func (proc readWriteCloseMock) Write(p []byte) (int, error) {
	if strings.HasPrefix(string(p), "Content-Length: ") {
		return 1, nil
	}
	var diagnostics jsonRpcDiagnostics
	json.NewDecoder(strings.NewReader(string(p))).Decode(&diagnostics)

	proc.diagnosticsChan <- diagnostics.Params
	return 1, nil
}

func (proc readWriteCloseMock) Close() error {
	return nil
}

func getYamllsConnector(t *testing.T, config util.YamllsConfiguration, customHandler *CustomHandler) (*Connector, *document.DocumentStore, chan lsp.PublishDiagnosticsParams) {
	dir := t.TempDir()
	documents := document.NewDocumentStore()
	diagnosticsChan := make(chan lsp.PublishDiagnosticsParams, 10000) // set a big size for tests where the channel is not read (prevents deadlock)
	con := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloseMock{diagnosticsChan}))
	zapLogger, _ := zap.NewProduction()
	client := protocol.ClientDispatcher(con, zapLogger)

	yamllsConnector := NewConnector(context.Background(), config, client, documents, customHandler)

	if yamllsConnector.server == nil {
		t.Fatal("Could not connect to yaml-language-server")
	}

	yamllsConnector.CallInitialize(context.Background(), uri.File(dir))

	return yamllsConnector, documents, diagnosticsChan
}

func openFile(t *testing.T, documents *document.DocumentStore, path string, yamllsConnector *Connector) {
	fileURI := uri.File(path)

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("Could not read test file", err)
	}
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       string(content),
		},
	}
	doc, err := documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
	yamllsConnector.DocumentDidOpenTemplate(doc.Ast, d)
}
