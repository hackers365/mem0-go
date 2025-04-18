package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bytectlgo/mem0-go/client"
	"github.com/bytectlgo/mem0-go/types"
)

var (
	host             string
	organizationName string
	projectName      string
	organizationID   string
	projectID        string
)

func init() {
	flag.StringVar(&host, "host", "https://api.mem0.ai", "Mem0 API host")
	flag.StringVar(&organizationName, "org-name", "", "Organization name")
	flag.StringVar(&projectName, "project-name", "", "Project name")
	flag.StringVar(&organizationID, "org-id", "", "Organization ID")
	flag.StringVar(&projectID, "project-id", "", "Project ID")
}

func main() {
	flag.Parse()

	apiKey := os.Getenv("MEM0_API_KEY")
	if apiKey == "" {
		log.Fatal("MEM0_API_KEY environment variable is required")
	}

	// 创建客户端
	mem0, err := client.NewMemoryClient(client.ClientOptions{
		APIKey:           apiKey,
		Host:             host,
		OrganizationName: organizationName,
		ProjectName:      projectName,
		OrganizationID:   organizationID,
		ProjectID:        projectID,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 根据命令行参数执行相应操作
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: mem0 <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  add <text> - Add a new memory")
		fmt.Println("  get <id> - Get a memory by ID")
		fmt.Println("  search <query> - Search memories")
		fmt.Println("  delete <id> - Delete a memory")
		return
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			log.Fatal("Text is required for add command")
		}
		memories, err := mem0.Add(args[1], types.MemoryOptions{
			UserID: "default_user",
		})
		if err != nil {
			log.Fatal(err)
		}
		printJSON(memories)

	case "get":
		if len(args) < 2 {
			log.Fatal("ID is required for get command")
		}
		memory, err := mem0.Get(args[1])
		if err != nil {
			log.Fatal(err)
		}
		printJSON(memory)

	case "search":
		if len(args) < 2 {
			log.Fatal("Query is required for search command")
		}
		results, err := mem0.Search(args[1], &types.SearchOptions{
			MemoryOptions: types.MemoryOptions{
				UserID: "default_user",
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		printJSON(results)

	case "delete":
		if len(args) < 2 {
			log.Fatal("ID is required for delete command")
		}
		err := mem0.Delete(args[1])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Memory deleted successfully")

	default:
		log.Fatalf("Unknown command: %s", args[0])
	}
}

func printJSON(v interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		log.Fatal(err)
	}
}
