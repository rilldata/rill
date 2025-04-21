import { markdown } from "@codemirror/lang-markdown";
import { sql } from "@codemirror/lang-sql";
import { yaml } from "@codemirror/lang-yaml";
import { type LanguageSupport } from "@codemirror/language";
import { DuckDBSQL } from "../../components/editor/presets/duckDBDialect";
import { extractFileExtension } from "../entity-management/file-path-utils";

export const FileExtensionToEditorExtension: Record<string, LanguageSupport[]> =
  {
    ".md": [markdown()],
    ".sql": [sql({ dialect: DuckDBSQL })],
    ".yaml": [yaml()],
    ".yml": [yaml()],
  };

export function getExtensionsForFile(filePath: string) {
  const extension = extractFileExtension(filePath);
  return FileExtensionToEditorExtension[extension] ?? [];
}
