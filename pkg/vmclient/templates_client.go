package vmclient

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

type CreateTemplateRequest struct {
	Name   string `json:"name"`
	Engine string `json:"engine"`
}

type TemplateDetailResponse struct {
	Template     *vmmodels.VM              `json:"template"`
	Settings     *vmmodels.VMSettings      `json:"settings,omitempty"`
	Capabilities []*vmmodels.VMCapability  `json:"capabilities"`
	StartupFiles []*vmmodels.VMStartupFile `json:"startup_files"`
}

func (c *Client) CreateTemplate(ctx context.Context, request CreateTemplateRequest) (*vmmodels.VM, error) {
	var template vmmodels.VM
	if err := c.do(ctx, "POST", "/api/v1/templates", request, &template); err != nil {
		return nil, err
	}
	return &template, nil
}

func (c *Client) ListTemplates(ctx context.Context) ([]*vmmodels.VM, error) {
	var templates []*vmmodels.VM
	if err := c.do(ctx, "GET", "/api/v1/templates", nil, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *Client) GetTemplate(ctx context.Context, templateID string) (*TemplateDetailResponse, error) {
	var response TemplateDetailResponse
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/templates/%s", templateID), nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) DeleteTemplate(ctx context.Context, templateID string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/templates/%s", templateID), nil, nil)
}

type AddTemplateCapabilityRequest struct {
	Kind    string      `json:"kind"`
	Name    string      `json:"name"`
	Enabled bool        `json:"enabled"`
	Config  interface{} `json:"config"`
}

func (c *Client) AddTemplateCapability(ctx context.Context, templateID string, request AddTemplateCapabilityRequest) (*vmmodels.VMCapability, error) {
	var capability vmmodels.VMCapability
	path := fmt.Sprintf("/api/v1/templates/%s/capabilities", templateID)
	if err := c.do(ctx, "POST", path, request, &capability); err != nil {
		return nil, err
	}
	return &capability, nil
}

func (c *Client) ListTemplateCapabilities(ctx context.Context, templateID string) ([]*vmmodels.VMCapability, error) {
	var capabilities []*vmmodels.VMCapability
	path := fmt.Sprintf("/api/v1/templates/%s/capabilities", templateID)
	if err := c.do(ctx, "GET", path, nil, &capabilities); err != nil {
		return nil, err
	}
	return capabilities, nil
}

type AddTemplateStartupFileRequest struct {
	Path       string `json:"path"`
	OrderIndex int    `json:"order_index"`
	Mode       string `json:"mode"`
}

type AddTemplateModuleRequest struct {
	Name string `json:"name"`
}

type AddTemplateLibraryRequest struct {
	Name string `json:"name"`
}

type TemplateNamedResourceResponse struct {
	TemplateID string `json:"template_id"`
	Name       string `json:"name"`
	Status     string `json:"status,omitempty"`
}

func (c *Client) AddTemplateStartupFile(ctx context.Context, templateID string, request AddTemplateStartupFileRequest) (*vmmodels.VMStartupFile, error) {
	var startup vmmodels.VMStartupFile
	path := fmt.Sprintf("/api/v1/templates/%s/startup-files", templateID)
	if err := c.do(ctx, "POST", path, request, &startup); err != nil {
		return nil, err
	}
	return &startup, nil
}

func (c *Client) ListTemplateStartupFiles(ctx context.Context, templateID string) ([]*vmmodels.VMStartupFile, error) {
	var startupFiles []*vmmodels.VMStartupFile
	path := fmt.Sprintf("/api/v1/templates/%s/startup-files", templateID)
	if err := c.do(ctx, "GET", path, nil, &startupFiles); err != nil {
		return nil, err
	}
	return startupFiles, nil
}

func (c *Client) ListTemplateModules(ctx context.Context, templateID string) ([]string, error) {
	var modules []string
	path := fmt.Sprintf("/api/v1/templates/%s/modules", templateID)
	if err := c.do(ctx, "GET", path, nil, &modules); err != nil {
		return nil, err
	}
	return modules, nil
}

func (c *Client) AddTemplateModule(ctx context.Context, templateID string, request AddTemplateModuleRequest) (*TemplateNamedResourceResponse, error) {
	var response TemplateNamedResourceResponse
	path := fmt.Sprintf("/api/v1/templates/%s/modules", templateID)
	if err := c.do(ctx, "POST", path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) RemoveTemplateModule(ctx context.Context, templateID, moduleName string) (*TemplateNamedResourceResponse, error) {
	var response TemplateNamedResourceResponse
	path := fmt.Sprintf("/api/v1/templates/%s/modules/%s", templateID, moduleName)
	if err := c.do(ctx, "DELETE", path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) ListTemplateLibraries(ctx context.Context, templateID string) ([]string, error) {
	var libraries []string
	path := fmt.Sprintf("/api/v1/templates/%s/libraries", templateID)
	if err := c.do(ctx, "GET", path, nil, &libraries); err != nil {
		return nil, err
	}
	return libraries, nil
}

func (c *Client) AddTemplateLibrary(ctx context.Context, templateID string, request AddTemplateLibraryRequest) (*TemplateNamedResourceResponse, error) {
	var response TemplateNamedResourceResponse
	path := fmt.Sprintf("/api/v1/templates/%s/libraries", templateID)
	if err := c.do(ctx, "POST", path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) RemoveTemplateLibrary(ctx context.Context, templateID, libraryName string) (*TemplateNamedResourceResponse, error) {
	var response TemplateNamedResourceResponse
	path := fmt.Sprintf("/api/v1/templates/%s/libraries/%s", templateID, libraryName)
	if err := c.do(ctx, "DELETE", path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
