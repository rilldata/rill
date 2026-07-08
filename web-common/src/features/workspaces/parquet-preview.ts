/**
 * Query key for the DuckDB-powered preview of a data file (e.g. .parquet)
 * rendered by ParquetWorkspace. It is keyed on the file's project-relative
 * path (the same form as file watch events) so the preview can be invalidated
 * when the underlying file is rewritten.
 */
export function getParquetPreviewQueryKey(instanceId: string, path: string) {
  return ["parquet-preview", instanceId, path];
}
