import type { JSONSchemaField } from "./schemas/types";

export type FileProcessingResult = {
  encodedContent: string;
  extractedValues: Record<string, unknown>;
};

export function processFileContent(
  content: string,
  field: JSONSchemaField,
): FileProcessingResult {
  const encoding = field["x-file-encoding"] ?? "raw";
  const extractMap = field["x-file-extract"];
  let encodedContent: string;
  let parsed: Record<string, unknown> | null = null;

  switch (encoding) {
    case "base64":
      try {
        encodedContent = btoa(content);
      } catch {
        throw new Error(
          "Invalid file encoding: contains non-Latin-1 characters",
        );
      }
      break;

    case "json":
      try {
        parsed = JSON.parse(content);
        encodedContent = JSON.stringify(parsed);
      } catch {
        throw new Error("Invalid JSON file");
      }
      break;

    case "raw":
    default:
      encodedContent = content;
      break;
  }

  const extractedValues: Record<string, unknown> = {};
  if (extractMap && parsed) {
    for (const [formFieldKey, jsonKey] of Object.entries(extractMap)) {
      const value = parsed[jsonKey];
      if (value !== undefined) {
        extractedValues[formFieldKey] = value;
      }
    }
  }

  return { encodedContent, extractedValues };
}

export function getFileAccept(field: JSONSchemaField): string | undefined {
  return field["x-file-accept"] ?? field["x-accept"];
}
