import { getFileTypeFromPath } from "../sources/sourceUtils";
import {
  FolderToResourceKind,
  ResourceShortNameToKind,
  removeLeadingSlash,
} from "./entity-mappers";
import { ResourceKind } from "./resource-selectors";

/**
 * When a file's reconciliation does not yield a resource, we guess the intended resource kind
 * in order to show the relevant Workspace UI.
 */
export function inferResourceKind(
  path: string,
  blob: string,
): ResourceKind | null {
  const fileType = getFileTypeFromPath(path);

  // If it's not a YAML or SQL file, we don't know what it is
  if (fileType !== "yaml" && fileType !== "sql") {
    return null;
  }

  // If it's a SQL file, it's a model
  if (fileType === "sql") {
    return ResourceKind.Model;
  }

  // If it is a YAML file...

  // Look at the file's content to see if it includes `type: <resource-kind>`
  const kindInText = findResourceKindInYAML(blob);
  if (kindInText) return kindInText;

  // Look at the first folder in the path
  const kindInPath = findResourceKindInFilePath(path);
  if (kindInPath) return kindInPath;

  return null;
}

function findResourceKindInYAML(text: string): ResourceKind | null {
  const regex = /type\s*:\s*(\w+)/i;
  const match = text.match(regex);

  if (match) {
    const shortName = match[1].toLowerCase();
    return ResourceShortNameToKind[shortName] ?? null;
  }

  return null;
}

function findResourceKindInFilePath(filePath: string): ResourceKind | null {
  const dirName = removeLeadingSlash(filePath).split("/")[0];
  return FolderToResourceKind[dirName] ?? null;
}
