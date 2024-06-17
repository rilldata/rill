import { FolderToResourceKind } from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  extractFileName,
  splitFolderAndName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  ResourceKind,
  ResourceShortNameToKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { parse } from "yaml";

export function parseKindAndNameFromFile(
  filePath: string,
  fileContents: string,
): V1ResourceName | undefined {
  const [folder, fileName] = splitFolderAndName(filePath);
  const kind = FolderToResourceKind[folder];
  const name = extractFileName(fileName);

  if (filePath.endsWith(".yaml") || filePath.endsWith(".yml")) {
    return tryParseYaml(kind, name, fileContents);
  } else if (filePath.endsWith(".sql")) {
    // .sql is defaulted to Model
    return tryParseSql(ResourceKind.Model, name, fileContents);
  }
  return undefined;
}

function tryParseYaml(
  kindFromFolder: ResourceKind | undefined,
  nameFromFolder: string,
  fileContents: string,
): V1ResourceName | undefined {
  let kind = kindFromFolder;
  let name = nameFromFolder;

  try {
    const yaml = parse(fileContents);

    // Get `type` (or `kind`, for backwards-compatibility) from yaml file
    // We try `kind` first to avoid picking up old Sources' `type` field
    if (yaml?.kind) {
      kind = ResourceShortNameToKind[yaml.kind as string];
    } else if (yaml?.type) {
      kind = ResourceShortNameToKind[yaml.type as string];
    }

    // Get `name` from yaml file
    if (yaml?.name) {
      name = yaml.name as string;
    }
  } catch (err) {
    const kindMatches = /^type\s*:\s*(.+?)\s*$/gm.exec(fileContents);
    if (kindMatches?.[1]) {
      kind = ResourceShortNameToKind[kindMatches?.[1] ?? ""];
    }
    const nameMatches = /^name\s*:\s*(.+?)\s*$/gm.exec(fileContents);
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
  nameFromFolder: string,
  fileContents: string,
): V1ResourceName | undefined {
  let kind = kindFromFolder;
  let name = nameFromFolder;

  const kindMatches = /^--\s*@type\s*:\s*(.+?)\s*$/gm.exec(fileContents);
  if (kindMatches?.[1]) {
    kind = ResourceShortNameToKind[kindMatches?.[1] ?? ""];
  }
  const nameMatches = /^--\s*@name\s*:\s*(.+?)\s*$/gm.exec(fileContents);
  if (nameMatches?.[1]) {
    name = nameMatches?.[1];
  }

  if (!kind) return undefined;
  return {
    kind,
    name,
  };
}
