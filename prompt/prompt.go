package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"uber-style-guide",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// create a prompt
	prompt := mcp.NewPrompt("uber-style-guide",

		mcp.WithPromptDescription("Review code for best practices and potential issues"),
		mcp.WithArgument("code",
			mcp.ArgumentDescription("code that follows uber style guide or not!"),
		),
	)

	s.AddPrompt(prompt, promptHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func promptHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	code := request.Params.Arguments["code"]

	guidePath := os.Getenv("STYLE_GUIDE_PATH")

	file, err := os.Open(guidePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open style guide: %w", err)
	}
	defer file.Close()

	var uberMarkdown string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		uberMarkdown += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read style guide: %w", err)
	}

	return mcp.NewGetPromptResult(
		"",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(fmt.Sprintf(
					"Please review this code to see if it follows Uber's Go Style Guide.\n\n"+
						"============= Uber Style Guide =============\n\n%s\n\n"+
						"============= Candidate Code =============\n\n%s",
					uberMarkdown, code,
				)),
			),
		},
	), nil
}
