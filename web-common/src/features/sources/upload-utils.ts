// TODO: Use this to block deploy
import type { JSONSchemaField } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";

export const UploadFileSizeLimitInBytes = 100 * 1024 * 1024; // 100MB limit

export function validateFileSize(
  files: unknown,
  prop: JSONSchemaField,
): string[] {
  if (prop["x-file-size-soft-limit"] === true) return [];
  const limit = prop["x-file-size-limit"];
  if (!limit) return [];
  if (!(files instanceof FileList)) return [];
  return Array.from(files)
    .map((file) => {
      if (file.size > limit) {
        return `File exceeds the maximum size of ${formatMemorySize(limit)}. Please choose a smaller file to continue.`;
      }
      return "";
    })
    .filter(Boolean);
}
