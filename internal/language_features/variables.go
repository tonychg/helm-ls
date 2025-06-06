package languagefeatures

import (
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type VariablesFeature struct {
	*GenericDocumentUseCase
}

func NewVariablesFeature(genericDocumentUseCase *GenericDocumentUseCase) *VariablesFeature {
	return &VariablesFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
	}
}

func (f *VariablesFeature) AppropriateForNode() bool {
	return (f.NodeType == gotemplate.NodeTypeIdentifier || f.NodeType == gotemplate.NodeTypeDollar) &&
		f.ParentNodeType == gotemplate.NodeTypeVariable
}

func (f *VariablesFeature) Definition() (result []lsp.Location, err error) {
	variableDefinition, err := f.Document.SymbolTable.GetVariableDefinitionForNode(f.GenericDocumentUseCase.Node, []byte(f.Document.Content))
	if err != nil {
		return []lsp.Location{}, err
	}

	return []lsp.Location{util.RangeToLocation(f.Document.URI, variableDefinition.Range)}, nil
}

func (f *VariablesFeature) References() (result []lsp.Location, err error) {
	variableReferences, err := f.Document.SymbolTable.GetVariableReferencesForNode(f.GenericDocumentUseCase.Node, []byte(f.Document.Content))
	if err != nil {
		return []lsp.Location{}, err
	}

	result = append(result, util.RangesToLocations(f.Document.URI, variableReferences)...)
	return result, nil
}

func (f *VariablesFeature) Completion() (result *lsp.CompletionList, err error) {
	return protocol.CompletionResults{}.WithVariableDefinitions(f.Document.SymbolTable.GetAllVariableDefinitions()).ToList(), nil
}
