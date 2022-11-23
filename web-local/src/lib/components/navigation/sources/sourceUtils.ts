import type { V1Connector } from "@rilldata/web-common/runtime-client";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { waitUntil } from "@rilldata/web-local/common/utils/waitUtils";
import type { PersistentTableStore } from "@rilldata/web-local/lib/application-state-stores/table-stores";
import { get } from "svelte/store";
import { sanitizeEntityName } from "../../../util/extract-table-name";

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

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string
) {
  if (connectorName !== "file") {
    values.uri = values.path;
    delete values.path;
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `type: "${connectorName}"\n` + compiledKeyValues;
}

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string
) {
  if (connectorName !== "file") {
    values.uri = values.path;
    delete values.path;
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `type: "${connectorName}"\n` + compiledKeyValues;
}

export function inferSourceName(connector: V1Connector, path: string) {
  if (
    !path ||
    path.endsWith("/") ||
    (connector.name === "gcs" && !path.startsWith("gs://")) ||
    (connector.name === "s3" && !path.startsWith("s3://")) ||
    (connector.name === "https" &&
      !path.startsWith("https://") &&
      !path.startsWith("http://"))
  )
    return;

  const slug = path
    .split("/")
    .filter((s: string) => s.length > 0)
    .pop();

  if (!slug) return;

  const fileName = slug.split(".").shift();

  if (!fileName) return;

  return sanitizeEntityName(fileName);
}
