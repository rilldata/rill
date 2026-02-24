export const FILES_WITHOUT_AUTOSAVE = ["/rill.yaml", "/.env"];

export const DIRECTORIES_WITHOUT_AUTOSAVE = ["/connectors", "/sources"];

/** Source YAML files now live in /models/ but should still require manual save. */
export function isFileWithoutAutosave(filePath: string): boolean {
  if (FILES_WITHOUT_AUTOSAVE.includes(filePath)) return true;
  const folderName = filePath.split("/").slice(0, -1).join("/");
  if (DIRECTORIES_WITHOUT_AUTOSAVE.includes(folderName)) return true;
  if (folderName === "/models" && filePath.endsWith(".yaml")) return true;
  return false;
}

export const FILE_SAVE_DEBOUNCE_TIME = 650;
