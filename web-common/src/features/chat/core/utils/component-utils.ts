// Helper to check if a tool result contains chart data
export function isChartToolResult(toolResult: any): boolean {
  if (!toolResult || !toolResult.content) return false;
  try {
    const parsed = JSON.parse(toolResult.content);
    return !!(parsed.chart_type && parsed.spec);
  } catch {
    return false;
  }
}
