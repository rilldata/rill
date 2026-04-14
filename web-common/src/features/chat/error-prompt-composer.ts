/**
 * Composes a minimal prompt for the AI developer agent to fix errors in a file.
 * The LLM agent can fetch parser errors and file content via tool calls.
 */
export function composeErrorPrompt(filePath: string): string {
  return `Fix the errors in \`${filePath}\``;
}
