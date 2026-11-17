package extensionimpl_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLoadExtensionDataValidation(t *testing.T) {
	managerConfig := []byte(`{"extensionName":"NNNN-gocfl-extension-manager","sort":{},"exclusion":{}}`)
	unknownMemberConfig := []byte(`{"extensionName":"NNNN-gocfl-extension-manager","unsupportedProperty":"value"}`)
	tests := []struct {
		name        string
		data        []byte
		params      map[string]string
		strict      *bool
		wantError   string
		wantLogged  bool
		wantContext string
	}{
		{name: "valid config is accepted", data: managerConfig},
		{name: "valid config is accepted in strict mode", data: managerConfig, params: map[string]string{extension.ParamStrictExtensionConfig: "true"}},
		{name: "unknown member is tolerated by default", data: unknownMemberConfig},
		{name: "unknown member is rejected by parameter", data: unknownMemberConfig, params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, wantError: "unsupportedProperty", wantLogged: true, wantContext: "provided extension config data"},
		{name: "unknown member is rejected by alias", data: unknownMemberConfig, params: map[string]string{extension.ParamValidateExtensionConfig: "true"}, wantError: "unsupportedProperty", wantLogged: true, wantContext: "provided extension config data"},
		{name: "factory option enables strict validation", data: unknownMemberConfig, strict: boolPointer(true), wantError: "unsupportedProperty", wantLogged: true, wantContext: "provided extension config data"},
		{name: "factory option disables strict validation", data: unknownMemberConfig, params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, strict: boolPointer(false)},
		{name: "invalid JSON syntax is rejected", data: []byte(`{"extensionName":"NNNN-gocfl-extension-manager"`), params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, wantError: "cannot unmarshal", wantLogged: true, wantContext: "provided extension config data"},
		{name: "missing extensionName is rejected", data: []byte(`{"sort":{}}`), params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, wantError: "extensionName", wantLogged: true, wantContext: "provided extension config data"},
		{name: "non-string extensionName is rejected", data: []byte(`{"extensionName":12}`), params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, wantError: "extensionName", wantLogged: true, wantContext: "provided extension config data"},
		{name: "incompatible config type is rejected", data: []byte(`{"extensionName":"NNNN-gocfl-extension-manager","sort":12}`), params: map[string]string{extension.ParamStrictExtensionConfig: "true"}, wantError: "sort", wantLogged: true, wantContext: "provided extension config data"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var zerologLogger zerolog.Logger
			logger := ocfl.NewOCFLLogger(t.Context(), &zerologLogger, nil, version.Version1_1, nil)
			factory, err := extensionimpl.NewFactory[object.ExtensionManager](test.params, logger)
			require.NoError(t, err)
			if test.strict != nil {
				factory = factory.WithStrictExtensionConfig(*test.strict)
			}

			_, err = factory.LoadExtensionData(test.data, nil)
			if test.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.wantError)
				if test.wantContext != "" {
					require.Contains(t, err.Error(), test.wantContext)
				}
			}
			if test.wantLogged {
				requireValidationDiagnostic(t, logger.ValidationErrors(), test.wantContext, test.wantError)
			}
		})
	}
}

func TestLoadExtensionFileValidationContext(t *testing.T) {
	var zerologLogger zerolog.Logger
	logger := ocfl.NewOCFLLogger(t.Context(), &zerologLogger, nil, version.Version1_1, nil)
	factory, err := extensionimpl.NewFactory[object.ExtensionManager](map[string]string{extension.ParamStrictExtensionConfig: "true"}, logger)
	require.NoError(t, err)

	fsys := fstest.MapFS{
		"config.json": &fstest.MapFile{Data: []byte(`{"extensionName":"NNNN-gocfl-extension-manager","unsupportedProperty":"value"}`)},
	}
	_, err = factory.LoadExtensionFile(fsys)
	require.ErrorContains(t, err, "config.json")
	require.ErrorContains(t, err, "unsupportedProperty")
	requireValidationDiagnostic(t, logger.ValidationErrors(), "config.json", "unsupportedProperty")
}

func TestLoadExtensionManagerValidationContextUsesFolderName(t *testing.T) {
	tests := []struct {
		name, folder, data, cause string
	}{
		{"unknown field", "unknown-name-a", `{"extensionName":"NNNN-gocfl-extension-manager","unsupportedProperty":"value"}`, "unsupportedProperty"},
		{"invalid field type", "unknown-name-b", `{"extensionName":"NNNN-gocfl-extension-manager","sort":12}`, "sort"},
		{"invalid syntax", "initial", `{"extensionName":`, "cannot unmarshal"},
		{"missing name", "missing-name", `{"sort":{}}`, "extensionName"},
		{"initial wrapper", "initial", `{"extensionName":"NNNN-gocfl-extension-manager","sort":{}}`, "sort"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var zerologLogger zerolog.Logger
			logger := ocfl.NewOCFLLogger(t.Context(), &zerologLogger, nil, version.Version1_1, nil)
			factory, err := extensionimpl.NewFactory[object.ExtensionManager](map[string]string{extension.ParamStrictExtensionConfig: "true"}, logger)
			require.NoError(t, err)

			fsys := fstest.MapFS{
				test.folder + "/config.json": &fstest.MapFile{Data: []byte(test.data)},
			}
			_, err = factory.LoadExtensionManager(fsys)
			configPath := "extensions/" + test.folder + "/config.json"
			require.ErrorContains(t, err, configPath)
			require.ErrorContains(t, err, test.cause)
			requireValidationDiagnostic(t, logger.ValidationErrors(), configPath, test.cause)
		})
	}
}

func requireValidationDiagnostic(t *testing.T, validationErrors []*validation.Error, context, cause string) {
	t.Helper()
	count := 0
	for _, validationError := range validationErrors {
		if validationError.Code != validation.E000 {
			continue
		}
		count++
		message := validationError.Error()
		require.True(t, strings.Contains(message, context) && strings.Contains(message, cause), "diagnostic %q should include %q and %q", message, context, cause)
	}
	require.Equal(t, 1, count, "expected one E000 diagnostic")
}

func boolPointer(value bool) *bool {
	return &value
}
