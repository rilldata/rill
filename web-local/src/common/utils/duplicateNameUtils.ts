import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { getName } from "@rilldata/web-local/common/utils/incrementName";

export function getDuplicateNameChecker(
  name: string,
  models: Array<PersistentModelEntity>,
  sources: Array<PersistentTableEntity>
) {
  return (
    models.some((model) => model.name === name) ||
    sources.some((source) => source.name === name)
  );
}

export function getIncrementedNameGetter(
  name: string,
  models: Array<PersistentModelEntity>,
  sources: Array<PersistentTableEntity>
) {
  return getName(name, [
    ...models.map((model) => model.name),
    ...sources.map((source) => source.name),
  ]);
}
