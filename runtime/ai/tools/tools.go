package tools

func newToolResult(result any, err error) map[string]any {
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}
	return map[string]any{
		"result": result,
	}
}
