import { EntityType } from "@rilldata/web-common/features/entity-management/types";

export function getFilePathFromPagePath(path: string): string {
  const pathSplits = path.split("/");
  const entityType = pathSplits[1];
  const entityName = pathSplits[2];

  switch (entityType) {
    case "source":
      return `sources/${entityName}.yaml`;
    case "model":
      return `models/${entityName}.sql`;
    case "dashboard":
      return `dashboards/${entityName}.yaml`;
    default:
      throw new Error("type must be either 'source', 'model', or 'dashboard'");
  }
}

export function getFilePathFromNameAndType(
  name: string,
  type: EntityType,
): string {
  switch (type) {
    case EntityType.Table:
      return `sources/${name}.yaml`;
    case EntityType.Model:
      return `models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `dashboards/${name}.yaml`;
    default:
      throw new Error(
        "type must be either 'Table', 'Model', or 'MetricsDefinition'",
      );
  }
}
// Temporary solution for the issue with leading `/` for these files.
// TODO: find a solution that works across backend and frontend
export function getFileAPIPathFromNameAndType(
  name: string,
  type: EntityType,
): string {
  switch (type) {
    case EntityType.Table:
      return `sources/${name}.yaml`;
    case EntityType.Model:
      return `models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `dashboards/${name}.yaml`;
    case EntityType.Chart:
      return `charts/${name}.yaml`;
    case EntityType.Dashboard:
      return `custom-dashboards/${name}.yaml`;
    default:
      throw new Error("Unrecognized EntityType");
  }
}

export function getNameFromFile(fileName: string): string {
  // TODO: do we need a library here?
  const splits = fileName.split("/");
  const extensionSplits = splits[splits.length - 1]?.split(".");
  return extensionSplits[0];
}

export function getRouteFromName(name: string, type: EntityType): string {
  if (!name) return "/";
  switch (type) {
    case EntityType.Table:
      return `/source/${name}`;
    case EntityType.Model:
      return `/model/${name}`;
    case EntityType.MetricsDefinition:
      return `/dashboard/${name}`;
    case EntityType.Chart:
      return `/chart/${name}`;
    case EntityType.Dashboard:
      return `/custom-dashboard/${name}`;
    default:
      throw new Error("Unrecognized EntityType");
  }
}

export function getLabel(entityType: EntityType) {
  switch (entityType) {
    case EntityType.Table:
      return "source";
    case EntityType.Model:
      return "model";
    case EntityType.MetricsDefinition:
      return "dashboard";
    case EntityType.Chart:
      return "chart";
    case EntityType.Dashboard:
      return "custom dashboard";
    default:
      throw new Error("Unrecognized EntityType");
  }
}

// Remove a leading slash, if it exists
export function removeLeadingSlash(path: string): string {
  return path.replace(/^\//, "");
}
