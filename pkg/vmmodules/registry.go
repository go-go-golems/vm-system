package vmmodules

import (
	"fmt"
	"sort"
	"strings"

	goja "github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	_ "github.com/go-go-golems/go-go-goja/modules/database"
	_ "github.com/go-go-golems/go-go-goja/modules/exec"
	_ "github.com/go-go-golems/go-go-goja/modules/fs"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

var jsBuiltinModuleSet = map[string]struct{}{
	"console": {},
	"math":    {},
	"json":    {},
	"date":    {},
	"array":   {},
	"string":  {},
	"object":  {},
	"promise": {},
}

func normalizeModuleName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// IsJSBuiltinModule reports whether the module name references a JavaScript
// built-in that should not be template-configurable.
func IsJSBuiltinModule(name string) bool {
	_, ok := jsBuiltinModuleSet[normalizeModuleName(name)]
	return ok
}

// RegisteredModuleNames returns sorted registered go-go-goja module names.
func RegisteredModuleNames() []string {
	docs := modules.DefaultRegistry.GetDocumentation()
	names := make([]string, 0, len(docs))
	for name := range docs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ValidateConfiguredModuleName validates template module input against the
// go-go-goja registry and policy constraints.
func ValidateConfiguredModuleName(name string) (string, error) {
	normalized := normalizeModuleName(name)
	if normalized == "" {
		return "", fmt.Errorf("%w: module name is required", vmmodels.ErrModuleNotAllowed)
	}
	if IsJSBuiltinModule(normalized) {
		return "", fmt.Errorf("%w: %q is a JavaScript built-in and cannot be configured per template", vmmodels.ErrModuleNotAllowed, normalized)
	}
	if modules.GetModule(normalized) == nil {
		return "", fmt.Errorf("%w: %q is not a registered native module", vmmodels.ErrModuleNotAllowed, normalized)
	}
	return normalized, nil
}

// RegisteredModuleLoaders returns Loader functions for the given module names,
// validated against the default registry. Returns an error if any name is invalid.
func RegisteredModuleLoaders(names []string) (map[string]func(*goja.Runtime, *goja.Object), error) {
	loaders := make(map[string]func(*goja.Runtime, *goja.Object))
	for _, rawName := range names {
		name, err := ValidateConfiguredModuleName(rawName)
		if err != nil {
			return nil, err
		}
		module := modules.GetModule(name)
		if module == nil {
			return nil, fmt.Errorf("%w: %q is not a registered native module", vmmodels.ErrModuleNotAllowed, name)
		}
		loaders[name] = module.Loader
	}
	return loaders, nil
}
