package devbrowser

// ToolExecutor defines how a tool should be executed
type ToolExecutor func(args map[string]any)

// BinaryData represents binary response data with metadata
// (Imported from mcpserve)
type BinaryData struct {
	MimeType string // e.g., "image/png", "application/pdf"
	Data     []byte // Raw binary data
}

// ToolMetadata provides MCP tool configuration metadata
type ToolMetadata struct {
	Name        string
	Description string
	Parameters  []ParameterMetadata
	Execute     ToolExecutor // Execution function
}

// ParameterMetadata describes a tool parameter
type ParameterMetadata struct {
	Name        string
	Description string
	Required    bool
	Type        string
	EnumValues  []string
	Default     any
}

// GetMCPToolsMetadata returns metadata for all DevBrowser MCP tools
func (b *DevBrowser) GetMCPToolsMetadata() []ToolMetadata {
	tools := []ToolMetadata{}
	tools = append(tools, b.getManagementTools()...)
	tools = append(tools, b.getConsoleTools()...)
	tools = append(tools, b.getScreenshotTools()...)
	//tools = append(tools, b.getEvaluateJsTools()...)
	//tools = append(tools, b.getNetworkTools()...)
	//tools = append(tools, b.getErrorTools()...)
	//tools = append(tools, b.getInteractionTools()...)
	return tools
}
