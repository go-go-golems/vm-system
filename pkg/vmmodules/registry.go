package vmmodules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	gogojamodules "github.com/go-go-golems/go-go-goja/modules"
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
	docs := gogojamodules.DefaultRegistry.GetDocumentation()
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
	if gogojamodules.GetModule(normalized) == nil {
		return "", fmt.Errorf("%w: %q is not a registered native module", vmmodels.ErrModuleNotAllowed, normalized)
	}
	return normalized, nil
}

// EnableConfiguredModules enables template-configured go-go-goja native
// modules and installs require() for the provided runtime.
func EnableConfiguredModules(vm *goja.Runtime, configured []string) error {
	reg := require.NewRegistry()
	seen := map[string]struct{}{}

	for _, rawName := range configured {
		name, err := ValidateConfiguredModuleName(rawName)
		if err != nil {
			return err
		}
		if _, ok := seen[name]; ok {
			continue
		}
		module := gogojamodules.GetModule(name)
		if module == nil {
			return fmt.Errorf("%w: %q is not a registered native module", vmmodels.ErrModuleNotAllowed, name)
		}
		reg.RegisterNativeModule(name, module.Loader)
		seen[name] = struct{}{}
	}

	reg.Enable(vm)
	return nil
}
