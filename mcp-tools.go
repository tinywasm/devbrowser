package devbrowser

import "github.com/tinywasm/mcp"

// GetMCPToolsMetadata returns metadata for all DevBrowser MCP tools
func (b *DevBrowser) GetMCPToolsMetadata() []mcp.ToolMetadata {
	tools := []mcp.ToolMetadata{}
	tools = append(tools, b.getManagementTools()...)
	tools = append(tools, b.getConsoleTools()...)
	tools = append(tools, b.getScreenshotTools()...)
	tools = append(tools, b.getStructureTools()...)
	//tools = append(tools, b.getEvaluateJsTools()...)
	//tools = append(tools, b.getNetworkTools()...)
	//tools = append(tools, b.getErrorTools()...)
	tools = append(tools, b.getInteractionTools()...)
	tools = append(tools, b.getNavigationTools()...)
	tools = append(tools, b.getInspectTools()...)
	tools = append(tools, b.getPerformanceTools()...)
	return tools
}
