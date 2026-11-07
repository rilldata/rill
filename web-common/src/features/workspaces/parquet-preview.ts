import type { PartialMessage, Struct } from "@bufbuild/protobuf";
import { createRuntimeServiceQueryResolver } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// Number of rows to preview. Parquet files can be large, so we cap the preview
// the same way the model results table does.
export const PARQUET_PREVIEW_LIMIT = 150;

/**
 * Query key for the DuckDB-powered preview of a Parquet file rendered by
 * ParquetWorkspace. It is keyed on the file's project-relative path (the same
 * form as file watch events) so the preview can be invalidated when the
 * underlying file is rewritten.
 */
export function getParquetPreviewQueryKey(instanceId: string, path: string) {
  return ["parquet-preview", instanceId, path];
}

/**
 * Creates the query that previews a Parquet file by reading it with DuckDB.
 *
 * `read_parquet` is DuckDB-specific and resolves relative paths against its own
 * working directory, so the caller must pass the file's absolute path and a
 * DuckDB connector to run the query against.
 */
export function createParquetPreviewQuery(
  client: RuntimeClient,
  {
    path,
    absolutePath,
    connector,
  }: { path: string; absolutePath: string; connector: string },
) {
  // Single quotes are escaped to keep the generated SQL valid.
  const sql = `SELECT * FROM read_parquet('${absolutePath.replaceAll("'", "''")}')`;
  return createRuntimeServiceQueryResolver(
    client,
    {
      resolver: "sql",
      resolverProperties: {
        connector,
        sql,
        limit: PARQUET_PREVIEW_LIMIT,
      } as unknown as PartialMessage<Struct>,
    },
    {
      query: {
        enabled: !!absolutePath && !!connector,
        queryKey: getParquetPreviewQueryKey(client.instanceId, path),
      },
    },
  );
}
