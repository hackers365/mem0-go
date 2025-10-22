package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hackers365/mem0-go/types"
)

// APIError 定义 API 错误
type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// ClientOptions 定义客户端选项
type ClientOptions struct {
	APIKey           string
	Host             string
	OrganizationName string
	ProjectName      string
	OrganizationID   string
	ProjectID        string
}

// MemoryClient 定义内存客户端
type MemoryClient struct {
	apiKey           string
	host             string
	organizationName string
	projectName      string
	organizationID   string
	projectID        string
	client           *http.Client
	telemetryID      string
}

// NewMemoryClient 创建新的内存客户端
func NewMemoryClient(options ClientOptions) (*MemoryClient, error) {
	if options.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	if options.Host == "" {
		options.Host = "https://api.mem0.ai"
	}

	client := &MemoryClient{
		apiKey:           options.APIKey,
		host:             options.Host,
		organizationName: options.OrganizationName,
		projectName:      options.ProjectName,
		organizationID:   options.OrganizationID,
		projectID:        options.ProjectID,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	if err := client.validateOrgProject(); err != nil {
		return nil, err
	}

	if err := client.ping(); err != nil {
		return nil, err
	}

	return client, nil
}

// validateOrgProject 验证组织和项目
func (c *MemoryClient) validateOrgProject() error {
	if (c.organizationName == "" && c.projectName != "") || (c.organizationName != "" && c.projectName == "") {
		return errors.New("both organizationName and projectName must be provided together")
	}

	if (c.organizationID == "" && c.projectID != "") || (c.organizationID != "" && c.projectID == "") {
		return errors.New("both organizationID and projectID must be provided together")
	}

	return nil
}

// ping 检查 API 连接
func (c *MemoryClient) ping() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/ping/", c.host), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: "API key is invalid"}
	}

	var data struct {
		Status    string `json:"status"`
		OrgID     string `json:"org_id"`
		ProjectID string `json:"project_id"`
		UserEmail string `json:"user_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if data.Status != "ok" {
		return &APIError{Message: "API key is invalid"}
	}

	c.organizationID = data.OrgID
	c.projectID = data.ProjectID
	c.telemetryID = data.UserEmail

	return nil
}

// preparePayload 准备请求体
func (c *MemoryClient) preparePayload(messages interface{}, options types.MemoryOptions) (map[string]interface{}, error) {
	payload := make(map[string]interface{})

	switch m := messages.(type) {
	case string:
		payload["messages"] = []types.Message{{Role: "user", Content: m}}
	case []types.Message:
		payload["messages"] = m
	default:
		return nil, errors.New("invalid messages type")
	}

	// 添加组织和项目信息
	if c.organizationName != "" && c.projectName != "" {
		options.OrgName = c.organizationName
		options.ProjectName = c.projectName
	}

	if c.organizationID != "" && c.projectID != "" {
		options.OrgID = c.organizationID
		options.ProjectID = c.projectID
	}

	// 将 options 转换为 map
	optionsMap := make(map[string]interface{})
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(optionsBytes, &optionsMap); err != nil {
		return nil, err
	}

	// 合并 payload 和 options
	for k, v := range optionsMap {
		if v != nil {
			payload[k] = v
		}
	}

	// 确保 messages 字段存在
	if _, ok := payload["messages"]; !ok {
		payload["messages"] = []types.Message{}
	}

	return payload, nil
}

// doRequest 执行 HTTP 请求
func (c *MemoryClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.host, path), reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	if c.telemetryID != "" {
		req.Header.Set("Mem0-User-ID", c.telemetryID)
	}

	return c.client.Do(req)
}

// Add 添加内存
func (c *MemoryClient) Add(messages interface{}, options types.MemoryOptions) ([]types.Memory, error) {
	payload, err := c.preparePayload(messages, options)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest("POST", "/v1/memories/", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取并打印响应体
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body))}
	}

	var memories []types.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Update 更新内存
func (c *MemoryClient) Update(memoryID string, message string) ([]types.Memory, error) {
	payload := map[string]string{
		"text": message,
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/v1/memories/%s/", memoryID), payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var memories []types.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Get 获取内存
func (c *MemoryClient) Get(memoryID string) (*types.Memory, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/v1/memories/%s/", memoryID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取并打印响应体
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body))}
	}

	var memory types.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// GetAll 获取所有内存
func (c *MemoryClient) GetAll(options *types.SearchOptions) ([]types.Memory, error) {
	path := "/v1/memories/"
	if options != nil {
		query := options.ToQuery()
		if query != "" {
			path += "?" + query
		}
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var memories []types.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Search 搜索内存
func (c *MemoryClient) Search(query string, options *types.SearchOptions) ([]types.Memory, error) {
	if options == nil {
		options = &types.SearchOptions{}
	}

	// 添加组织和项目信息
	if c.organizationName != "" && c.projectName != "" {
		options.OrgName = c.organizationName
		options.ProjectName = c.projectName
	}

	if c.organizationID != "" && c.projectID != "" {
		options.OrgID = c.organizationID
		options.ProjectID = c.projectID
	}

	// 准备请求体
	payload := map[string]interface{}{
		"query": query,
	}

	// 将 options 转换为 map
	optionsMap := make(map[string]interface{})
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(optionsBytes, &optionsMap); err != nil {
		return nil, err
	}

	// 合并 payload 和 options
	for k, v := range optionsMap {
		if v != nil {
			payload[k] = v
		}
	}

	resp, err := c.doRequest("POST", "/v1/memories/search/", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取并打印响应体
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body))}
	}

	var memories []types.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// Delete 删除内存
func (c *MemoryClient) Delete(memoryID string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/v1/memories/%s/", memoryID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取并打印响应体
		body, _ := io.ReadAll(resp.Body)
		return &APIError{Message: fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body))}
	}

	return nil
}

// DeleteAll 删除所有内存
func (c *MemoryClient) DeleteAll(options types.MemoryOptions) error {
	path := "/v1/memories/"
	if query := options.ToQuery(); query != "" {
		path += "?" + query
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// History 获取内存历史
func (c *MemoryClient) History(memoryID string) ([]types.MemoryHistory, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/v1/memories/%s/history/", memoryID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var history []types.MemoryHistory
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}

	return history, nil
}

// Users 获取所有用户
func (c *MemoryClient) Users() (*types.AllUsers, error) {
	resp, err := c.doRequest("GET", "/v1/users/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var users types.AllUsers
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return &users, nil
}

// DeleteUser 删除用户
func (c *MemoryClient) DeleteUser(entityID string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/v1/users/%s/", entityID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// DeleteUsers 删除所有用户
func (c *MemoryClient) DeleteUsers() error {
	resp, err := c.doRequest("DELETE", "/v1/users/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// BatchUpdate 批量更新内存
func (c *MemoryClient) BatchUpdate(memories []types.MemoryUpdateBody) error {
	resp, err := c.doRequest("PUT", "/v1/memories/batch/", memories)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// BatchDelete 批量删除内存
func (c *MemoryClient) BatchDelete(memoryIDs []string) error {
	resp, err := c.doRequest("DELETE", "/v1/memories/batch/", memoryIDs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// GetProject 获取项目
func (c *MemoryClient) GetProject(options types.ProjectOptions) (*types.ProjectResponse, error) {
	path := "/v1/project/"
	if query := options.ToQuery(); query != "" {
		path += "?" + query
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var project types.ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &project, nil
}

// UpdateProject 更新项目
func (c *MemoryClient) UpdateProject(payload types.PromptUpdatePayload) error {
	resp, err := c.doRequest("PUT", "/v1/project/", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// GetWebhooks 获取 Webhooks
func (c *MemoryClient) GetWebhooks(projectID string) ([]types.Webhook, error) {
	path := "/v1/webhooks/"
	if projectID != "" {
		path += "?project_id=" + projectID
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var webhooks []types.Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, err
	}

	return webhooks, nil
}

// CreateWebhook 创建 Webhook
func (c *MemoryClient) CreateWebhook(webhook types.WebhookPayload) (*types.Webhook, error) {
	resp, err := c.doRequest("POST", "/v1/webhooks/", webhook)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	var createdWebhook types.Webhook
	if err := json.NewDecoder(resp.Body).Decode(&createdWebhook); err != nil {
		return nil, err
	}

	return &createdWebhook, nil
}

// UpdateWebhook 更新 Webhook
func (c *MemoryClient) UpdateWebhook(webhook types.WebhookPayload) error {
	resp, err := c.doRequest("PUT", "/v1/webhooks/", webhook)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// DeleteWebhook 删除 Webhook
func (c *MemoryClient) DeleteWebhook(webhookID string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/v1/webhooks/%s/", webhookID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}

// Feedback 提交反馈
func (c *MemoryClient) Feedback(payload types.FeedbackPayload) error {
	resp, err := c.doRequest("POST", "/v1/feedback/", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &APIError{Message: fmt.Sprintf("API request failed with status %d", resp.StatusCode)}
	}

	return nil
}
