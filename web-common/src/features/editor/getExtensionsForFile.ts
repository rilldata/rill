import { markdown } from "@codemirror/lang-markdown";
import { yaml } from "@codemirror/lang-yaml";
import { type LanguageSupport } from "@codemirror/language";
import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";

export const FileExtensionToEditorExtension: Record<string, LanguageSupport[]> =
  {
    ".yaml": [yaml()],
    ".yml": [yaml()],
    ".md": [markdown()],
  };

export function getExtensionsForFile(filePath: string) {
  const extension = extractFileExtension(filePath);
  return FileExtensionToEditorExtension[extension] ?? [];
}
