package charts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestSetValuesFilesConfigOverwrites(t *testing.T) {
	valuesFilesConfig := util.ValuesFilesConfig{
		MainValuesFileName:               "value.yaml",
		AdditionalValuesFilesGlobPattern: "values*.yaml",
		LintOverlayValuesFileName:        "something.yaml",
	}
	tempDir := t.TempDir()
	valuesContent := `foo: bar`

	_ = os.WriteFile(filepath.Join(tempDir, "Chart.yaml"), []byte(""), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "value.yaml"), []byte("foo: main"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "something.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values.other.yaml"), []byte(valuesContent), 0o644)
	s := NewChartStore(uri.New(util.FileURIScheme+tempDir), NewChart)

	chart, err := s.GetChartForURI(uri.New(util.FileURIScheme + tempDir))
	assert.Equal(t, chartutil.Values{}, chart.ValuesFiles.MainValuesFile.Values)

	s.SetValuesFilesConfig(valuesFilesConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, chartutil.Values{}, chart.ValuesFiles.MainValuesFile.Values)
	assert.Equal(t, valuesFilesConfig.MainValuesFileName, filepath.Base(chart.ValuesFiles.MainValuesFile.URI.Filename()))
	assert.Equal(t, valuesFilesConfig.LintOverlayValuesFileName, filepath.Base(chart.ValuesFiles.OverlayValuesFile.URI.Filename()))
}

func TestSetValuesFilesConfigDoesNotOverwrite(t *testing.T) {
	valuesFilesConfig := util.ValuesFilesConfig{
		MainValuesFileName:               "values.yaml",
		AdditionalValuesFilesGlobPattern: "values*.yaml",
		LintOverlayValuesFileName:        "values.lint.yaml",
	}
	tempDir := t.TempDir()
	valuesContent := `foo: bar`

	_ = os.WriteFile(filepath.Join(tempDir, "Chart.yaml"), []byte(""), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte("foo: main"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "something.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values.lint.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values.other.yaml"), []byte(valuesContent), 0o644)
	s := NewChartStore(uri.New(util.FileURIScheme+tempDir), NewChart)

	chart, err := s.GetChartForURI(uri.New(util.FileURIScheme + tempDir))
	assert.NoError(t, err)
	assert.NotEqual(t, chartutil.Values{}, chart.ValuesFiles.MainValuesFile.Values)

	s.SetValuesFilesConfig(valuesFilesConfig)
	chart, err = s.GetChartForURI(uri.New(util.FileURIScheme + tempDir))
	assert.NoError(t, err)
	assert.Equal(t, valuesFilesConfig.MainValuesFileName, filepath.Base(chart.ValuesFiles.MainValuesFile.URI.Filename()))
	assert.Equal(t, valuesFilesConfig.LintOverlayValuesFileName, filepath.Base(chart.ValuesFiles.OverlayValuesFile.URI.Filename()))
}

func TestGetChartForURIWhenChartYamlDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte("foo: main"), 0o644)
	s := NewChartStore(uri.New(util.FileURIScheme+tempDir), NewChart)

	chart, err := s.GetChartForURI(uri.New(util.FileURIScheme + tempDir))
	assert.Error(t, err)
	assert.Nil(t, chart)
}
