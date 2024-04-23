import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";

export function isCurrentActivePage(
  filePath: string,
  currentFile: string | undefined,
  isDir: boolean,
) {
  filePath = removeLeadingSlash(filePath);
  currentFile = removeLeadingSlash(currentFile ?? "");
  return (
    currentFile !== "" &&
    (currentFile === filePath || (isDir && filePath.startsWith(currentFile)))
  );
}
