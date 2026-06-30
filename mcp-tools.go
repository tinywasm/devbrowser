package devbrowser

import "github.com/tinywasm/mcp"

// GetMCPTools returns metadata for all DevBrowser MCP tools
func (b *DevBrowser) GetMCPTools() []mcp.Tool {
	tools := []mcp.Tool{}
	tools = append(tools, b.GetManagementTools()...)
	tools = append(tools, b.GetConsoleTools()...)
	tools = append(tools, b.GetScreenshotTools()...)
	tools = append(tools, b.GetStructureTools()...)
	tools = append(tools, b.GetEvaluateJsTools()...)
	tools = append(tools, b.GetNetworkTools()...)
	tools = append(tools, b.GetErrorTools()...)
	tools = append(tools, b.GetInteractionTools()...)
	tools = append(tools, b.GetNavigationTools()...)
	tools = append(tools, b.GetInspectTools()...)
	tools = append(tools, b.GetPerformanceTools()...)
	tools = append(tools, b.GetSourceTools()...)
	tools = append(tools, b.GetStylesTools()...)
	tools = append(tools, b.GetStorageTools()...)
	tools = append(tools, b.GetAssetTools()...)
	tools = append(tools, b.GetInterceptTools()...)
	return tools
}
