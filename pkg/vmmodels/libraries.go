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

// ExposedModule represents a host module exposed to the VM
type ExposedModule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Kind        string            `json:"kind"` // "builtin", "host", "custom"
	Description string            `json:"description"`
	Functions   []string          `json:"functions"` // List of exposed functions
	Config      map[string]string `json:"config"`
}

// BuiltinModules returns a list of built-in modules that can be exposed to VMs
func BuiltinModules() []ExposedModule {
	return []ExposedModule{
		{
			ID:          "console",
			Name:        "console",
			Kind:        "builtin",
			Description: "Console logging and debugging",
			Functions:   []string{"log", "warn", "error", "info", "debug"},
			Config:      map[string]string{},
		},
		{
			ID:          "math",
			Name:        "Math",
			Kind:        "builtin",
			Description: "Mathematical functions and constants",
			Functions:   []string{"abs", "ceil", "floor", "round", "sqrt", "pow", "random"},
			Config:      map[string]string{},
		},
		{
			ID:          "json",
			Name:        "JSON",
			Kind:        "builtin",
			Description: "JSON parsing and stringification",
			Functions:   []string{"parse", "stringify"},
			Config:      map[string]string{},
		},
		{
			ID:          "date",
			Name:        "Date",
			Kind:        "builtin",
			Description: "Date and time manipulation",
			Functions:   []string{"now", "parse", "UTC"},
			Config:      map[string]string{},
		},
		{
			ID:          "array",
			Name:        "Array",
			Kind:        "builtin",
			Description: "Array manipulation methods",
			Functions:   []string{"map", "filter", "reduce", "forEach", "find", "some", "every"},
			Config:      map[string]string{},
		},
		{
			ID:          "string",
			Name:        "String",
			Kind:        "builtin",
			Description: "String manipulation methods",
			Functions:   []string{"split", "join", "slice", "substring", "indexOf", "replace"},
			Config:      map[string]string{},
		},
		{
			ID:          "object",
			Name:        "Object",
			Kind:        "builtin",
			Description: "Object manipulation methods",
			Functions:   []string{"keys", "values", "entries", "assign", "freeze"},
			Config:      map[string]string{},
		},
		{
			ID:          "promise",
			Name:        "Promise",
			Kind:        "builtin",
			Description: "Asynchronous programming with promises",
			Functions:   []string{"resolve", "reject", "all", "race"},
			Config:      map[string]string{},
		},
	}
}
