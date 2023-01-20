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
