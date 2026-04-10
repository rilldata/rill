/**
 * Composes a structured prompt from error context for the AI developer agent.
 */

const MAX_PROMPT_LENGTH = 2000;
const SHORT_FILE_THRESHOLD = 50;
const CONTEXT_LINES_ABOVE = 5;
const CONTEXT_LINES_BELOW = 5;
const LONG_FILE_HEAD = 30;
const LONG_FILE_TAIL = 10;

export interface ErrorPromptInput {
  errorMessage: string;
  filePath: string;
  fileContent?: string | null;
  lineNumber?: number;
  additionalErrorCount?: number;
}

export function composeErrorPrompt(input: ErrorPromptInput): string {
  const {
    errorMessage,
    filePath,
    fileContent,
    lineNumber,
    additionalErrorCount,
  } = input;

  const resourceType = inferResourceType(filePath);
  const isConnector = resourceType === "connector";

  const parts: string[] = [];

  // Header
  parts.push(
    `I have an error in my ${resourceType} file \`${filePath}\`${lineNumber ? ` at Line ${lineNumber}` : ""}:`,
  );
  parts.push("");

  // Error message
  parts.push(`**Error:** ${errorMessage}`);
  parts.push("");

  // File context; track indices for potential truncation
  let codeBlockStart = -1;
  let codeBlockEnd = -1;

  if (fileContent) {
    const sanitized = isConnector ? stripCredentials(fileContent) : fileContent;
    const context = extractContext(sanitized, lineNumber);
    if (context) {
      codeBlockStart = parts.length;
      parts.push("**Relevant code:**");
      parts.push("```");
      parts.push(context);
      parts.push("```");
      codeBlockEnd = parts.length;
      parts.push("");
    }
  }

  // Additional errors
  if (additionalErrorCount && additionalErrorCount > 0) {
    parts.push(
      `There are also ${additionalErrorCount} other errors in this file.`,
    );
    parts.push("");
  }

  // Closing
  parts.push("Please explain what's wrong and suggest how to fix it.");

  let prompt = parts.join("\n");

  // Truncate if over limit; drop code block by index range
  if (prompt.length > MAX_PROMPT_LENGTH && codeBlockStart >= 0) {
    const withoutCode = parts
      .filter((_, i) => i < codeBlockStart || i >= codeBlockEnd)
      .join("\n");
    if (withoutCode.length < MAX_PROMPT_LENGTH) {
      prompt = withoutCode;
    }
  }

  return prompt;
}

function inferResourceType(filePath: string): string {
  // Check directory-based types first (more specific than extension)
  if (filePath.startsWith("/models/")) return "SQL model";
  if (filePath.startsWith("/sources/")) return "source";
  if (filePath.startsWith("/connectors/")) return "connector";
  if (filePath.startsWith("/metrics/")) return "metrics view";
  if (filePath.startsWith("/dashboards/")) return "canvas dashboard";
  if (filePath.startsWith("/explores/")) return "explore dashboard";
  if (filePath.startsWith("/apis/")) return "API";
  if (filePath.startsWith("/themes/")) return "theme";

  // Fall back to extension-based detection
  if (filePath.endsWith(".sql")) return "SQL model";
  if (filePath.endsWith(".yaml") || filePath.endsWith(".yml")) return "YAML";
  return "file";
}

function stripCredentials(content: string): string {
  const pattern =
    /^(\s*)(password|secret|token|api_key|apikey|api_secret|access_key|private_key|client_secret|auth_token|credentials|connection_string)(\s*:\s*)(.+)$/gim;
  return content.replace(pattern, "$1$2$3[REDACTED]");
}

function extractContext(
  content: string,
  lineNumber: number | undefined,
): string {
  const lines = content.split("\n");

  if (lineNumber) {
    const idx = lineNumber - 1; // 0-based
    const start = Math.max(0, idx - CONTEXT_LINES_ABOVE);
    const end = Math.min(lines.length, idx + CONTEXT_LINES_BELOW + 1);
    return lines
      .slice(start, end)
      .map((line, i) => {
        const num = start + i + 1;
        const marker = num === lineNumber ? " >" : "  ";
        return `${marker} ${num}: ${line}`;
      })
      .join("\n");
  }

  if (lines.length <= SHORT_FILE_THRESHOLD) {
    return lines.map((line, i) => `  ${i + 1}: ${line}`).join("\n");
  }

  // Long file: first N + last M
  const head = lines
    .slice(0, LONG_FILE_HEAD)
    .map((line, i) => `  ${i + 1}: ${line}`);
  const tail = lines
    .slice(-LONG_FILE_TAIL)
    .map((line, i) => `  ${lines.length - LONG_FILE_TAIL + i + 1}: ${line}`);
  return [...head, "  ...", ...tail].join("\n");
}
