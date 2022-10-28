import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { waitUntil } from "@rilldata/web-local/common/utils/waitUtils";
import type { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores";
import { get } from "svelte/store";

export async function waitForSource(
  newTableName: string,
  store: PersistentTableStore
): Promise<string> {
  let foundTable: PersistentTableEntity;
  await waitUntil(() => {
    foundTable = get(store).entities.find(
      (source) => source.tableName.toLowerCase() === newTableName.toLowerCase()
    );
    return foundTable !== undefined;
  });
  return foundTable?.id;
}

export function compileCreateSourceSql(
  values: Record<string, unknown>,
  connectorName: string
) {
  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `'${key}'='${value}'`)
    .join(", ");

  return (
    `CREATE SOURCE ${values.sourceName} WITH (connector = '${connectorName}', ` +
    compiledKeyValues +
    `)`
  );
}
