import type { EntityRecord } from "@rilldata/web-local/lib/temp/entity";

export function getNextEntityId(
  entities: Array<EntityRecord>,
  entityId: string
): string {
  if (entities.length === 1) return undefined;
  const idx = entities.findIndex((entity) => entity.id === entityId);
  if (idx === 0) {
    return entities[idx + 1].id;
  } else {
    return entities[idx - 1].id;
  }
}

export function getNextEntityName(
  entityNames: Array<string>,
  entityName: string
): string {
  const idx = entityNames.indexOf(entityName);
  if (idx <= 0) {
    return entityNames[idx + 1];
  } else {
    return entityNames[idx - 1];
  }
}
