import { FolderToResourceKind } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  ResourceShortNameToKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { parse } from "yaml";

export function parseKindAndNameFromFile(
  filePath: string,
  fileContents: string,
): V1ResourceName | undefined {
  const [folder, fileName] = splitFolderAndName(filePath);
  const kind = FolderToResourceKind[folder];
  const name = extractFileName(fileName);

  if (filePath.endsWith(".yaml")) {
    return tryParseYaml(kind, name, fileContents);
  } else if (filePath.endsWith(".sql")) {
    return tryParseSql(kind, name, fileContents);
  }
  return undefined;
}

function tryParseYaml(
  kindFromFolder: ResourceKind | undefined,
  kindFromName: string,
  fileContents: string,
): V1ResourceName | undefined {
  let kind = kindFromFolder;
  let name = kindFromName;

  try {
    const yaml = parse(fileContents);
    if (yaml.kind) {
      kind = ResourceShortNameToKind[yaml.kind as string];
    }
    if (yaml.name) {
      name = yaml.name as string;
    }
  } catch (err) {
    const kindMatches = /^kind\s*:\s*(.*)\s*$/gm.exec(fileContents);
    if (kindMatches?.[1]) {
      kind = ResourceShortNameToKind[kindMatches?.[1] ?? ""];
    }
    const nameMatches = /^name\s*:\s*(.*)\s*$/gm.exec(fileContents);
    if (nameMatches?.[1]) {
      name = nameMatches?.[1];
    }
  }

  if (!kind) return undefined;
  return {
    kind,
    name,
  };
}

function tryParseSql(
  kindFromFolder: ResourceKind | undefined,
  kindFromName: string,
  fileContents: string,
): V1ResourceName | undefined {
  let kind = kindFromFolder;
  let name = kindFromName;

  const kindMatches = /^--\s*@kind\s*:\s*(.*)\s*$/gm.exec(fileContents);
  if (kindMatches?.[1]) {
    kind = ResourceShortNameToKind[kindMatches?.[1] ?? ""];
  }
  const nameMatches = /^--\s*@name\s*:\s*(.*)\s*$/gm.exec(fileContents);
  if (nameMatches?.[1]) {
    name = nameMatches?.[1];
  }

  if (!kind) return undefined;
  return {
    kind,
    name,
  };
}
