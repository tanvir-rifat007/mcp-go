package main

import (
	"context"
	"fmt"

	"github.com/hectormalot/omgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server

	s := server.NewMCPServer(
		"add-go-server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add tool
	tool := mcp.NewTool("add-go",
		mcp.WithDescription("Add two numbers"),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	// weather tool
	tool2 := mcp.NewTool("weather-go",
		mcp.WithDescription("Get current weather data for a given latitude and longitude"),
		mcp.WithNumber("lat",
			mcp.Required(),
			mcp.Description("Latitude coordinate"),
		),
		mcp.WithNumber("lon",
			mcp.Required(),
			mcp.Description("Longitude coordinate"),
		),
	)

	// Add tool handler
	s.AddTool(tool, addHandler)

	s.AddTool(tool2, weatherHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func addHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	x, err := request.RequireFloat("x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	y, err := request.RequireFloat("y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result := x + y
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil

}

func weatherHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lat, err := request.RequireFloat("lat")
	if err != nil {
		return mcp.NewToolResultError("missing lat"), nil
	}
	lon, err := request.RequireFloat("lon")
	if err != nil {
		return mcp.NewToolResultError("missing lon"), nil
	}

	// Create Open-Meteo client
	client, err := omgo.NewClient()
	if err != nil {
		return mcp.NewToolResultError("failed to create weather client"), nil
	}

	// Location from params
	loc, _ := omgo.NewLocation(lat, lon)

	// Fetch current weather
	res, err := client.CurrentWeather(context.Background(), loc, nil)
	if err != nil {
		return mcp.NewToolResultError("failed to fetch weather"), nil
	}

	// Return result as text
	output := fmt.Sprintf(
		"Weather at (%.2f, %.2f): %.1f°C, Windspeed %.1f km/h",
		lat, lon, res.Temperature, res.WindSpeed,
	)

	return mcp.NewToolResultText(output), nil
}
