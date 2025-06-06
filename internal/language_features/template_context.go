package languagefeatures

import (
	"fmt"
	"reflect"
	"strings"

	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/charts"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	"helm.sh/helm/v3/pkg/chart"
)

type TemplateContextFeature struct {
	*GenericTemplateContextFeature
}

func NewTemplateContextFeature(genericDocumentUseCase *GenericDocumentUseCase) *TemplateContextFeature {
	return &TemplateContextFeature{
		GenericTemplateContextFeature: &GenericTemplateContextFeature{genericDocumentUseCase},
	}
}

func (f *TemplateContextFeature) AppropriateForNode() bool {
	if f.NodeType == gotemplate.NodeTypeDot || f.NodeType == gotemplate.NodeTypeDotSymbol {
		return true
	}
	return (f.ParentNodeType == gotemplate.NodeTypeField && f.NodeType == gotemplate.NodeTypeIdentifier) ||
		f.NodeType == gotemplate.NodeTypeFieldIdentifier ||
		f.NodeType == gotemplate.NodeTypeField
}

func (f *TemplateContextFeature) References() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil || len(templateContext) == 0 {
		return []lsp.Location{}, err
	}

	locations := f.getReferencesFromSymbolTable(templateContext)
	return append(locations, f.getDefinitionLocations(templateContext)...), nil
}

func (f *TemplateContextFeature) Definition() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil || len(templateContext) == 0 {
		return []lsp.Location{}, err
	}
	return f.getDefinitionLocations(templateContext), nil
}

func (f *TemplateContextFeature) getDefinitionLocations(templateContext symboltable.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}

	switch templateContext[0] {
	case "Values":
		for _, value := range f.Chart.ResolveValueFiles(templateContext.Tail(), f.ChartStore) {
			locs := value.ValuesFiles.GetPositionsForValue(value.Selector)
			if len(locs) > 0 {
				for _, valuesFile := range value.ValuesFiles.AllValuesFiles() {
					charts.SyncToDisk(valuesFile)
				}
			}
			locations = append(locations, locs...)
		}
		return locations
	case "Chart":
		location, _ := f.Chart.GetMetadataLocation(templateContext.Tail())
		return []lsp.Location{location}
	}
	return locations
}

func (f *TemplateContextFeature) Hover() (string, error) {
	templateContext, err := f.getTemplateContext()
	if err != nil || len(templateContext) == 0 {
		return "", err
	}

	switch templateContext[0] {
	case "Values":
		return f.valuesHover(templateContext.Tail())
	case "Chart":
		docs, err := f.builtInOjectDocsLookup(templateContext.Tail().Format(), helmdocs.BuiltInOjectVals[templateContext[0]])
		value := f.getMetadataField(&f.Chart.ChartMetadata.Metadata, docs.Name)
		return fmt.Sprintf("%s\n\n%s\n", docs.Doc, value), err
	case "Release", "Files", "Capabilities", "Template":
		docs, err := f.builtInOjectDocsLookup(templateContext.Tail().Format(), helmdocs.BuiltInOjectVals[templateContext[0]])
		return docs.Doc, err
	}

	return templateContext.Format(), err
}

func (f *TemplateContextFeature) valuesHover(templateContext symboltable.TemplateContext) (string, error) {
	var (
		valuesFiles  = f.Chart.ResolveValueFiles(templateContext, f.ChartStore)
		hoverResults = protocol.HoverResultsWithFiles{}
	)
	for _, valuesFiles := range valuesFiles {
		for _, valuesFile := range valuesFiles.ValuesFiles.AllValuesFiles() {
			logger.Debug(fmt.Sprintf("Looking for selector: %s in values %v", strings.Join(valuesFiles.Selector, "."), valuesFile.Values))
			result, err := util.GetTableOrValueForSelector(valuesFile.Values, valuesFiles.Selector)

			if err == nil {
				hoverResults = append(hoverResults, protocol.HoverResultWithFile{URI: valuesFile.URI, Value: result})
			}
		}
	}
	return hoverResults.FormatYaml(f.ChartStore.RootURI), nil
}

func (f *TemplateContextFeature) getMetadataField(v *chart.Metadata, fieldName string) string {
	r := reflect.ValueOf(v)
	field := reflect.Indirect(r).FieldByName(fieldName)
	return util.FormatToYAML(field, fieldName)
}
