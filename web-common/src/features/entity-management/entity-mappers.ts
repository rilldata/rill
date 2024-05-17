import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";

export function getFilePathFromPagePath(path: string): string {
  const pathSplits = path.split("/");
  const entityType = pathSplits[1];
  const entityName = pathSplits[2];

  switch (entityType) {
    case "source":
      return `/files/sources/${entityName}.yaml`;
    case "model":
      return `/files/models/${entityName}.sql`;
    case "dashboard":
      return `/files/dashboards/${entityName}.yaml`;
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
      return `/sources/${name}.yaml`;
    case EntityType.Model:
      return `/models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `/dashboards/${name}.yaml`;
    case EntityType.Chart:
      return `/charts/${name}.yaml`;
    case EntityType.Dashboard:
      return `/custom-dashboards/${name}.yaml`;
    default:
      throw new Error("Unrecognized EntityType");
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
      return `/files/source/${name}`;
    case EntityType.Model:
      return `/files/model/${name}`;
    case EntityType.MetricsDefinition:
      return `/files/dashboard/${name}`;
    case EntityType.Chart:
      return `/files/chart/${name}`;
    case EntityType.Dashboard:
      return `/files/custom-dashboard/${name}`;
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
    case EntityType.Unknown:
      return "";
    default:
      throw new Error("Unrecognized EntityType");
  }
}

// Remove a leading slash, if it exists
// In certain backend APIs where path is part of the url.
// Leading slash leads to issues where the end point redirects to one without the slash.
export function removeLeadingSlash(path: string): string {
  return path.replace(/^\//, "");
}

// Add a leading slash if it doesn't exist.
// Temporary, we should eventually make sure this is added in all places
export function addLeadingSlash(path: string): string {
  if (path.startsWith("/")) return path;
  return "/" + path;
}

export const FolderToResourceKind: Record<string, ResourceKind> = {
  sources: ResourceKind.Source,
  models: ResourceKind.Model,
  dashboards: ResourceKind.MetricsView,
  charts: ResourceKind.Component,
  "custom-dashboards": ResourceKind.Dashboard,
};
