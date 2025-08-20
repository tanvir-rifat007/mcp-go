package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hectormalot/omgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sashabaranov/go-openai"
)

func main() {
	s := server.NewMCPServer(
		"weather-image-server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// weather tool
	tool1 := mcp.NewTool(
		"weather",

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

	// genearate an image tool based on that weather
	tool2 := mcp.NewTool("weather-image",
		mcp.WithDescription(" Fetch current weather at lat/lon and generate an AI image of the conditions"),

		mcp.WithNumber("lat", mcp.Required(), mcp.Description("Latitude coordinate")),
		mcp.WithNumber("lon", mcp.Required(), mcp.Description("Longitude coordinate")),
	)

	s.AddTool(tool1, weatherHandler)

	s.AddTool(tool2, weatherImageHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
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

func weatherImageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lat, err := request.RequireFloat("lat")
	if err != nil {
		return mcp.NewToolResultError("missing lat"), nil
	}
	lon, err := request.RequireFloat("lon")
	if err != nil {
		return mcp.NewToolResultError("missing lon"), nil
	}

	// Weather first
	client, _ := omgo.NewClient()
	loc, _ := omgo.NewLocation(lat, lon)
	res, err := client.CurrentWeather(context.Background(), loc, nil)
	if err != nil {
		return mcp.NewToolResultError("failed to fetch weather"), nil
	}

	// Build prompt
	prompt := fmt.Sprintf("Generate an artistic image of the location (%.2f, %.2f). The weather is %.1f°C with %.1f km/h wind.", lat, lon, res.Temperature, res.WindSpeed)

	// OpenAI image client
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	imgReq := openai.ImageRequest{
		Prompt:         prompt,
		Size:           "512x512",
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	imgResp, err := openaiClient.CreateImage(ctx, imgReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("image generation failed: %v", err)), nil
	}

	// Return the image URL
	return mcp.NewToolResultText(fmt.Sprintf("Weather: %.1f°C, %.1f km/h wind\nImage: %s", res.Temperature, res.WindSpeed, imgResp.Data[0].URL)), nil
}
