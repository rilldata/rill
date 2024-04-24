import { page } from "$app/stores";
import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { get } from "svelte/store";

export function isCurrentActivePage(filePath: string, isDir: boolean) {
  let currentFile = get(page).params.file;
  if (currentFile === undefined) return false; // handle case where user is on home page

  filePath = removeLeadingSlash(filePath);
  currentFile = removeLeadingSlash(currentFile ?? "");
  return (
    currentFile === filePath || (isDir && filePath.startsWith(currentFile))
  );
}
