import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { RESOURCE_FILE_EXTENSIONS } from "./file-path-utils";

export function getFilePathFromPagePath(path: string): string {
  const pathSplits = path.split("/");
  const entityType = pathSplits[1];
  const entityName = pathSplits[2];

  switch (entityType) {
    case "source":
      return `/files/models/${entityName}.yaml`;
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
    case EntityType.Connector:
      return `/connectors/${name}.yaml`;
    case EntityType.Table:
      return `/models/${name}.yaml`;
    case EntityType.Model:
      return `/models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `/dashboards/${name}.yaml`;
    case EntityType.Chart:
      return `/charts/${name}.yaml`;
    case EntityType.Canvas:
      return `/canvas/${name}.yaml`;
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
    case EntityType.Connector:
      return `connectors/${name}.yaml`;
    case EntityType.Table:
      return `models/${name}.yaml`;
    case EntityType.Model:
      return `models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `dashboards/${name}.yaml`;
    case EntityType.Chart:
      return `charts/${name}.yaml`;
    case EntityType.Canvas:
      return `canvas/${name}.yaml`;
    default:
      throw new Error("Unrecognized EntityType");
  }
}

export function getNameFromFile(fileName: string): string {
  const splits = fileName.split("/");
  const basename = splits[splits.length - 1] ?? "";

  // Rill resource names are inferred by removing only the final resource file
  // extension, so dotted names like `dashboard.canvas.yaml` stay intact.
  for (const extension of RESOURCE_FILE_EXTENSIONS) {
    if (basename.endsWith(extension)) {
      return basename.slice(0, -extension.length);
    }
  }

  // Non-resource data files keep the legacy behavior of removing compound
  // extensions, e.g. `adbids.csv.tgz` -> `adbids`.
  const extensionSplits = basename.split(".");
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
    case EntityType.Canvas:
      return `/files/canvas/${name}`;
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
    case EntityType.Canvas:
      return "canvas";
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
// While all files returned by Repo APIs have this already, AI interactions might not always have it.
export function addLeadingSlash(path: string): string {
  if (path.startsWith("/")) return path;
  return "/" + path;
}

export const FolderNameToResourceKind: Record<string, ResourceKind> = {
  sources: ResourceKind.Source,
  models: ResourceKind.Model,
  // dashboards: ResourceKind.MetricsView,
  metrics: ResourceKind.MetricsView,
  // explores: ResourceKind.Explore, // This does not happen on backend
  charts: ResourceKind.Component,
  canvas: ResourceKind.Canvas,
};

export const ResourceShortNameToResourceKind: Record<string, ResourceKind> = {
  alert: ResourceKind.Alert,
  api: ResourceKind.API,
  component: ResourceKind.Component,
  connector: ResourceKind.Connector,
  canvas: ResourceKind.Canvas,
  metrics_view: ResourceKind.MetricsView,
  metricsview: ResourceKind.MetricsView,
  explore: ResourceKind.Explore,
  model: ResourceKind.Model,
  report: ResourceKind.Report,
  source: ResourceKind.Source,
  theme: ResourceKind.Theme,
};
