package main

import (
	"context"
	"fmt"
	"os"

	"github.com/daveroberts0321/cloudpact/generator"
	"github.com/daveroberts0321/cloudpact/project"
	"github.com/daveroberts0321/cloudpact/watch"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact init <project-name>")
			return
		}
		projectName := os.Args[2]
		if err := project.Init(projectName); err != nil {
			fmt.Printf("Error initializing project: %v\n", err)
			return
		}
		fmt.Printf("Project '%s' initialized successfully!\n", projectName)
		fmt.Printf("   cd %s\n", projectName)
		fmt.Printf("   cloudpact start http\n")

	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact start <http|build>")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "http":
			if err := project.StartDevServer(); err != nil {
				fmt.Printf("Error starting dev server: %v\n", err)
			}
		case "build":
			if err := project.Build(); err != nil {
				fmt.Printf("Error building project: %v\n", err)
			} else {
				fmt.Println("Project built successfully!")
			}
		default:
			fmt.Printf("Unknown start command: %s\n", subCmd)
		}

	case "gen":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact gen <record|function|openapi> [args...]")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "record":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen record <RecordName>")
				return
			}
			generator.GenerateRecord(os.Args[3])
		case "function":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen function <FunctionName>")
				return
			}
			generator.GenerateFunction(os.Args[3])
		case "model":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen model <ModelName>")
				return
			}
			generator.GenerateModel(os.Args[3])
		case "openapi":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen openapi <file.cp>")
				return
			}
			if err := generator.GenerateOpenAPI(os.Args[3]); err != nil {
				fmt.Printf("Error generating OpenAPI: %v\n", err)
			}
		default:
			fmt.Printf("Unknown gen command: %s\n", subCmd)
		}

	case "ai":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact ai <review|feedback|status|accept> [args...]")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "review":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact ai review <file.cp>")
				return
			}
			fileName := os.Args[3]
			fmt.Printf("AI review for %s (not yet implemented)\n", fileName)
		case "feedback":
			fmt.Println("AI feedback session (not yet implemented)")
		case "status":
			fmt.Println("AI suggestions status (not yet implemented)")
		case "accept":
			fmt.Println("Accept AI suggestion (not yet implemented)")
		default:
			fmt.Printf("Unknown ai command: %s\n", subCmd)
		}

	case "watch":
		if err := watch.Watch(context.Background(), project.Build); err != nil {
			fmt.Printf("Error watching files: %v\n", err)
		}

	case "version":
		fmt.Println("CloudPact v0.2.0 - Human/AI collaborative programming language")

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`CloudPact - Human/AI collaborative programming language

USAGE:
    cloudpact <command> [arguments]

COMMANDS:
    init <name>           Initialize a new CloudPact project
    start http            Start development server with hot reload
    start build           Build the project once
    gen record <name>     Generate a record template
    gen function <name>   Generate a function template
    gen model <name>      Generate a model template (legacy)
    gen openapi <file>    Generate OpenAPI spec from .cp file
    ai review <file>      AI reviews a specific file
    ai feedback           Interactive AI feedback session
    ai status             Show pending AI suggestions
    ai accept <id>        Accept a specific AI suggestion
    watch                 Watch files and rebuild on changes
    version               Show version information
    help                  Show this help message

EXAMPLES:
    cloudpact init myapp
    cloudpact start http
    cloudpact gen record User
    cloudpact gen function validateUser
    cloudpact ai review models/user.cp
    cloudpact gen openapi models/user.cp`)
}
