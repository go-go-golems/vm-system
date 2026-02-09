package vmmodels

// Library represents a JavaScript library that can be loaded into a VM
type Library struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Source      string            `json:"source"` // URL or local path to library source
	Type        string            `json:"type"`   // "builtin", "npm", "url", "local"
	Config      map[string]string `json:"config"` // Library-specific configuration
}

// BuiltinLibraries returns a list of built-in libraries available for VMs
func BuiltinLibraries() []Library {
	return []Library{
		{
			ID:          "lodash",
			Name:        "Lodash",
			Version:     "4.17.21",
			Description: "A modern JavaScript utility library delivering modularity, performance & extras",
			Source:      "https://cdn.jsdelivr.net/npm/lodash@4.17.21/lodash.min.js",
			Type:        "npm",
			Config:      map[string]string{"global": "_"},
		},
		{
			ID:          "moment",
			Name:        "Moment.js",
			Version:     "2.29.4",
			Description: "Parse, validate, manipulate, and display dates and times in JavaScript",
			Source:      "https://cdn.jsdelivr.net/npm/moment@2.29.4/moment.min.js",
			Type:        "npm",
			Config:      map[string]string{"global": "moment"},
		},
		{
			ID:          "axios",
			Name:        "Axios",
			Version:     "1.6.0",
			Description: "Promise based HTTP client for the browser and node.js",
			Source:      "https://cdn.jsdelivr.net/npm/axios@1.6.0/dist/axios.min.js",
			Type:        "npm",
			Config:      map[string]string{"global": "axios"},
		},
		{
			ID:          "ramda",
			Name:        "Ramda",
			Version:     "0.29.0",
			Description: "A practical functional library for JavaScript programmers",
			Source:      "https://cdn.jsdelivr.net/npm/ramda@0.29.0/dist/ramda.min.js",
			Type:        "npm",
			Config:      map[string]string{"global": "R"},
		},
		{
			ID:          "dayjs",
			Name:        "Day.js",
			Version:     "1.11.10",
			Description: "Fast 2kB alternative to Moment.js with the same modern API",
			Source:      "https://cdn.jsdelivr.net/npm/dayjs@1.11.10/dayjs.min.js",
			Type:        "npm",
			Config:      map[string]string{"global": "dayjs"},
		},
		{
			ID:          "zustand",
			Name:        "Zustand",
			Version:     "4.4.7",
			Description: "A small, fast and scalable bearbones state-management solution",
			Source:      "https://cdn.jsdelivr.net/npm/zustand@4.4.7/index.js",
			Type:        "npm",
			Config:      map[string]string{"global": "zustand"},
		},
	}
}

// ExposedModule represents a configurable host module exposed to the VM runtime.
type ExposedModule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Kind        string            `json:"kind"` // "native", "host", "custom"
	Description string            `json:"description"`
	Functions   []string          `json:"functions"` // List of exposed functions
	Config      map[string]string `json:"config"`
}

// BuiltinModules returns the catalog of template-configurable native modules.
//
// Note: JavaScript language built-ins (JSON, Math, Date, etc.) are always
// available and are intentionally not configurable per template.
func BuiltinModules() []ExposedModule {
	return []ExposedModule{
		{
			ID:          "database",
			Name:        "database",
			Kind:        "native",
			Description: "SQLite database access from JavaScript via go-go-goja native module",
			Functions:   []string{"configure", "query", "exec", "close"},
			Config:      map[string]string{},
		},
		{
			ID:          "exec",
			Name:        "exec",
			Kind:        "native",
			Description: "Run external commands from JavaScript",
			Functions:   []string{"run"},
			Config:      map[string]string{},
		},
		{
			ID:          "fs",
			Name:        "fs",
			Kind:        "native",
			Description: "Read/write files from JavaScript via native module APIs",
			Functions:   []string{"readFileSync", "writeFileSync"},
			Config:      map[string]string{},
		},
	}
}
