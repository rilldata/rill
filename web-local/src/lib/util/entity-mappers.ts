import { EntityType } from "@rilldata/web-local/lib/temp/entity";

export function getFilePathFromNameAndType(
  name: string,
  type: EntityType
): string {
  switch (type) {
    case EntityType.Table:
      return `/sources/${name}.yaml`;
    case EntityType.Model:
      return `/models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `/dashboards/${name}.yaml`;
    default:
      throw new Error(
        "type must be either 'Table', 'Model', or 'MetricsDefinition'"
      );
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
    default:
      throw new Error(
        "type must be either 'Table', 'Model', or 'MetricsDefinition'"
      );
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
    default:
      throw new Error(
        "type must be either 'Table', 'Model', or 'MetricsDefinition'"
      );
  }
}
