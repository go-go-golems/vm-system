package vmclient

import "context"

func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := c.do(ctx, "GET", "/api/v1/health", nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) RuntimeSummary(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := c.do(ctx, "GET", "/api/v1/runtime/summary", nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}
