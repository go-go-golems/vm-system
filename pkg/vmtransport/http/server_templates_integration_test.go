package vmhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

func TestTemplateEndpointsCRUDAndNestedResources(t *testing.T) {
	server, client := newIntegrationTestServer(t)
	defer server.Close()

	template := struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Engine string `json:"engine"`
	}{}
	postJSON(t, client, server.URL+"/api/v1/templates", map[string]interface{}{
		"name": "template-endpoint-test",
	}, &template)

	if template.ID == "" {
		t.Fatalf("expected created template id")
	}
	if template.Engine != "goja" {
		t.Fatalf("expected default engine goja, got %q", template.Engine)
	}

	listResp := []struct {
		ID string `json:"id"`
	}{}
	getJSON(t, client, server.URL+"/api/v1/templates", &listResp)
	if len(listResp) == 0 {
		t.Fatalf("expected template list to include created template")
	}

	cases := []struct {
		name string
		path string
		body map[string]interface{}
	}{
		{
			name: "add capability",
			path: fmt.Sprintf("/api/v1/templates/%s/capabilities", template.ID),
			body: map[string]interface{}{
				"kind":    "module",
				"name":    "console",
				"enabled": true,
				"config":  map[string]interface{}{},
			},
		},
		{
			name: "add startup file",
			path: fmt.Sprintf("/api/v1/templates/%s/startup-files", template.ID),
			body: map[string]interface{}{
				"path":        "runtime/startup.js",
				"order_index": 10,
				"mode":        "eval",
			},
		},
		{
			name: "add module",
			path: fmt.Sprintf("/api/v1/templates/%s/modules", template.ID),
			body: map[string]interface{}{
				"name": "console",
			},
		},
		{
			name: "add library",
			path: fmt.Sprintf("/api/v1/templates/%s/libraries", template.ID),
			body: map[string]interface{}{
				"name": "lodash-4.17.21",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out map[string]interface{}
			postJSON(t, client, server.URL+tc.path, tc.body, &out)
		})
	}

	detail := struct {
		Template struct {
			ID             string   `json:"id"`
			ExposedModules []string `json:"exposed_modules"`
			Libraries      []string `json:"libraries"`
		} `json:"template"`
		Settings     map[string]interface{} `json:"settings"`
		Capabilities []struct {
			Name string `json:"name"`
		} `json:"capabilities"`
		StartupFiles []struct {
			Path string `json:"path"`
		} `json:"startup_files"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s", server.URL, template.ID), &detail)
	if detail.Template.ID != template.ID {
		t.Fatalf("expected template detail for %s, got %s", template.ID, detail.Template.ID)
	}
	if len(detail.Settings) == 0 {
		t.Fatalf("expected template settings in detail response")
	}
	if len(detail.Capabilities) != 1 || detail.Capabilities[0].Name != "console" {
		t.Fatalf("expected one console capability in detail response")
	}
	if len(detail.StartupFiles) != 1 || detail.StartupFiles[0].Path != "runtime/startup.js" {
		t.Fatalf("expected one startup file in detail response")
	}
	if len(detail.Template.ExposedModules) != 1 || detail.Template.ExposedModules[0] != "console" {
		t.Fatalf("expected one console module in detail response")
	}
	if len(detail.Template.Libraries) != 1 || detail.Template.Libraries[0] != "lodash-4.17.21" {
		t.Fatalf("expected one lodash library in detail response")
	}

	capabilities := []struct {
		Name string `json:"name"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/capabilities", server.URL, template.ID), &capabilities)
	if len(capabilities) != 1 {
		t.Fatalf("expected one capability, got %d", len(capabilities))
	}

	startupFiles := []struct {
		Path string `json:"path"`
	}{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/startup-files", server.URL, template.ID), &startupFiles)
	if len(startupFiles) != 1 {
		t.Fatalf("expected one startup file, got %d", len(startupFiles))
	}

	modules := []string{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, template.ID), &modules)
	if len(modules) != 1 || modules[0] != "console" {
		t.Fatalf("expected modules list [console], got %#v", modules)
	}

	libraries := []string{}
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/libraries", server.URL, template.ID), &libraries)
	if len(libraries) != 1 || libraries[0] != "lodash-4.17.21" {
		t.Fatalf("expected libraries list [lodash-4.17.21], got %#v", libraries)
	}

	doRequest(t, client, http.MethodDelete, fmt.Sprintf("%s/api/v1/templates/%s/modules/%s", server.URL, template.ID, "console"), nil, http.StatusOK, nil)
	doRequest(t, client, http.MethodDelete, fmt.Sprintf("%s/api/v1/templates/%s/libraries/%s", server.URL, template.ID, "lodash-4.17.21"), nil, http.StatusOK, nil)

	modules = nil
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/modules", server.URL, template.ID), &modules)
	if len(modules) != 0 {
		t.Fatalf("expected no modules after delete, got %#v", modules)
	}

	libraries = nil
	getJSON(t, client, fmt.Sprintf("%s/api/v1/templates/%s/libraries", server.URL, template.ID), &libraries)
	if len(libraries) != 0 {
		t.Fatalf("expected no libraries after delete, got %#v", libraries)
	}

	doRequest(t, client, http.MethodDelete, fmt.Sprintf("%s/api/v1/templates/%s", server.URL, template.ID), nil, http.StatusOK, nil)
	doRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/templates/%s", server.URL, template.ID), nil, http.StatusNotFound, map[string]string{
		"code": "TEMPLATE_NOT_FOUND",
	})
}

func newIntegrationTestServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vm-system.db")

	store, err := vmstore.NewVMStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	core := vmcontrol.NewCore(store)
	server := httptest.NewServer(vmhttp.NewHandler(core))
	return server, server.Client()
}

func doRequest(t *testing.T, client *http.Client, method, url string, in interface{}, expectedStatus int, expectedErr map[string]string) {
	t.Helper()

	var bodyReader *bytes.Reader
	if in != nil {
		body, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		raw, _ := ioReadAll(resp)
		t.Fatalf("expected status %d, got %d (%s)", expectedStatus, resp.StatusCode, string(raw))
	}

	if expectedErr != nil {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if envelope.Error.Code != expectedErr["code"] {
			t.Fatalf("expected error code %q, got %q", expectedErr["code"], envelope.Error.Code)
		}
	}
}
