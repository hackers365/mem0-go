# Mem0 Go Client

Mem0 Go 客户端是一个用于与 Mem0 API 交互的 Go 语言客户端库。

## 安装

```bash
go get github.com/bytectlgo/mem0-go
```

## 版本要求

- Go 1.18 或更高版本

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/bytectlgo/mem0-go/client"
	"github.com/bytectlgo/mem0-go/types"
)

func main() {
	// 创建客户端
	mem0, err := client.NewMemoryClient(client.ClientOptions{
		APIKey: "your-api-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 添加内存
	memories, err := mem0.Add("Hello, World!", types.MemoryOptions{
		UserID: "user-123",
		Metadata: map[string]interface{}{
			"source": "example",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Added memory: %+v\n", memories[0])

	// 搜索内存
	results, err := mem0.Search("Hello", &types.SearchOptions{
		Limit: 10,
		Threshold: 0.8,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Search results: %+v\n", results)

	// 获取项目信息
	project, err := mem0.GetProject(types.ProjectOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Project info: %+v\n", project)
}
```

## 功能特性

- 内存管理
  - 添加内存
  - 更新内存
  - 获取内存
  - 搜索内存
  - 删除内存
  - 批量操作
- 用户管理
  - 获取用户列表
  - 删除用户
- 项目管理
  - 获取项目信息
  - 更新项目设置
- Webhook 管理
  - 创建 Webhook
  - 更新 Webhook
  - 删除 Webhook
  - 获取 Webhook 列表
- 反馈
  - 提交反馈

## API 文档

### 客户端初始化

```go
client, err := client.NewMemoryClient(client.ClientOptions{
	APIKey:          "your-api-key",
	Host:            "https://api.mem0.ai", // 可选，默认为 https://api.mem0.ai
	OrganizationName: "org-name",           // 可选
	ProjectName:     "project-name",        // 可选
	OrganizationID:  "org-id",             // 可选
	ProjectID:       "project-id",         // 可选
})
```

### 内存操作

#### 添加内存

```go
memories, err := client.Add("Hello, World!", types.MemoryOptions{
	UserID: "user-123",
	Metadata: map[string]interface{}{
		"source": "example",
	},
})
```

#### 更新内存

```go
memories, err := client.Update("memory-id", "Updated content")
```

#### 获取内存

```go
memory, err := client.Get("memory-id")
```

#### 搜索内存

```go
results, err := client.Search("query", &types.SearchOptions{
	Limit: 10,
	Threshold: 0.8,
})
```

#### 删除内存

```go
err := client.Delete("memory-id")
```

### 用户管理

#### 获取用户列表

```go
users, err := client.Users()
```

#### 删除用户

```go
err := client.DeleteUser("user-id")
```

### 项目管理

#### 获取项目信息

```go
project, err := client.GetProject(types.ProjectOptions{})
```

#### 更新项目设置

```go
err := client.UpdateProject(types.PromptUpdatePayload{
	CustomInstructions: "New instructions",
})
```

### Webhook 管理

#### 创建 Webhook

```go
webhook, err := client.CreateWebhook(types.WebhookPayload{
	EventTypes: []types.WebhookEvent{types.MemoryAdded},
	ProjectID:  "project-id",
	Name:       "My Webhook",
	URL:        "https://example.com/webhook",
})
```

#### 更新 Webhook

```go
err := client.UpdateWebhook(types.WebhookPayload{
	WebhookID:  "webhook-id",
	EventTypes: []types.WebhookEvent{types.MemoryAdded, types.MemoryUpdated},
	Name:       "Updated Webhook",
	URL:        "https://example.com/webhook",
})
```

#### 删除 Webhook

```go
err := client.DeleteWebhook("webhook-id")
```

#### 获取 Webhook 列表

```go
webhooks, err := client.GetWebhooks("project-id")
```

### 反馈

```go
err := client.Feedback(types.FeedbackPayload{
	MemoryID:      "memory-id",
	Feedback:      types.Positive,
	FeedbackReason: "Helpful response",
})
```

## 错误处理

所有 API 方法都可能返回错误。错误类型包括：

- `APIError`: API 请求失败时返回
- 其他标准 Go 错误

```go
memory, err := client.Get("memory-id")
if err != nil {
	if apiErr, ok := err.(*client.APIError); ok {
		fmt.Printf("API Error: %s\n", apiErr.Message)
	} else {
		fmt.Printf("Error: %v\n", err)
	}
	return
}
```

## 许可证

MIT 

## 常见问题

### 如何处理 API 限流？

当遇到 API 限流时，建议实现指数退避重试机制。示例代码：

```go
func retryWithBackoff(fn func() error) error {
    var err error
    for i := 0; i < 3; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 429 {
            time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
            continue
        }
        return err
    }
    return err
}
```

### 如何批量处理内存？

使用 `AddBatch` 方法可以批量添加内存：

```go
memories := []string{"memory1", "memory2", "memory3"}
results, err := client.AddBatch(memories, types.MemoryOptions{})
```

## 贡献指南

我们欢迎任何形式的贡献！在提交 Pull Request 之前，请确保：

1. 代码符合 Go 标准格式
2. 添加了必要的测试
3. 更新了相关文档
4. 提交信息清晰明了

## 变更日志

### v0.1.0 (2025-04-18)
- 初始版本发布
- 支持基本的内存管理功能
- 支持用户管理功能
- 支持项目管理功能
- 支持 Webhook 管理功能
- 支持反馈功能 