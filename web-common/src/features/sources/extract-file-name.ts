import { extractFileName } from "../entity-management/file-path-utils";
import { sanitizeEntityName } from "../entity-management/name-utils";

export function getTableNameFromFile(filePath: string, name?: string) {
  return name ?? sanitizeEntityName(extractFileName(filePath));
}
