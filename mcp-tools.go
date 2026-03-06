package devbrowser

import "github.com/tinywasm/mcp"

// GetMCPTools returns metadata for all DevBrowser MCP tools
func (b *DevBrowser) GetMCPTools() []mcp.Tool {
	tools := []mcp.Tool{}
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
