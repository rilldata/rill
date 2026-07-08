<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServiceQueryResolver,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getParquetPreviewQueryKey } from "./parquet-preview";

  // Number of rows to preview. Parquet files can be large, so we cap the
  // preview the same way the model results table does.
  const PREVIEW_LIMIT = 150;

  let { fileArtifact }: { fileArtifact: FileArtifact } = $props();

  const runtimeClient = useRuntimeClient();

  let { path } = $derived(fileArtifact);
  let fileName = $derived(fileArtifact.fileName);

  let instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });

  // DuckDB resolves relative paths against its own working directory, not the
  // project root, so we must read the file by its absolute path. The project's
  // local root is the `dsn` of the repo connector (only present in Rill
  // Developer, where the project lives on the local filesystem).
  let repoRoot = $derived.by(() => {
    const instance = $instanceQuery.data?.instance;
    const repo = instance?.connectors?.find(
      (c) => c.name === instance?.repoConnector,
    );
    const dsn = repo?.config?.dsn;
    return typeof dsn === "string" ? dsn : "";
  });

  // read_parquet is DuckDB-specific, so the query must run against a DuckDB
  // connector even when the project's default OLAP connector is something else
  // (e.g. ClickHouse). Prefer the default OLAP connector if it is DuckDB,
  // otherwise fall back to any DuckDB connector in the project.
  let duckDbConnector = $derived.by(() => {
    const instance = $instanceQuery.data?.instance;
    const connectors = instance?.connectors ?? [];
    const olap = connectors.find((c) => c.name === instance?.olapConnector);
    if (olap?.type === "duckdb") return olap.name ?? "";
    return connectors.find((c) => c.type === "duckdb")?.name ?? "";
  });

  // Join the repo root and the file's project-relative path into an absolute
  // path. Single quotes are escaped to keep the generated SQL valid.
  let absolutePath = $derived(
    repoRoot
      ? `${repoRoot.replace(/\/+$/, "")}/${path.replace(/^\/+/, "")}`
      : "",
  );
  let sql = $derived(
    `SELECT * FROM read_parquet('${absolutePath.replaceAll("'", "''")}')`,
  );

  let previewQuery = $derived(
    createRuntimeServiceQueryResolver(
      runtimeClient,
      {
        resolver: "sql",
        resolverProperties: {
          connector: duckDbConnector,
          sql,
          limit: PREVIEW_LIMIT,
        } as unknown as PartialMessage<Struct>,
      },
      {
        query: {
          enabled: !!absolutePath && !!duckDbConnector,
          queryKey: getParquetPreviewQueryKey(runtimeClient.instanceId, path),
        },
      },
    ),
  );

  // While the instance is loading we don't yet know the repo root, so treat it
  // as part of the preview's loading state.
  let isLoading = $derived($instanceQuery.isLoading || $previewQuery.isLoading);
  let error = $derived.by(() => {
    if ($instanceQuery.isLoading) return null;
    if (!repoRoot)
      return new Error(
        "Could not resolve the project directory for this file.",
      );
    if (!duckDbConnector)
      return new Error("Previewing Parquet files requires a DuckDB connector.");
    return $previewQuery.error;
  });
  let data = $derived($previewQuery.data);

  let rows = $derived(data?.data ?? []);
  let columns = $derived(
    (data?.schema?.fields?.map((field) => ({
      name: field.name,
      type: field.type?.code,
    })) ?? []) as VirtualizedTableColumns[],
  );
</script>

<svelte:head>
  <title>Rill Developer | {fileName}</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    slot="header"
    filePath={path}
    resourceKind={undefined}
    titleInput={fileName}
    editable={false}
    showInspectorToggle={false}
    hasUnsavedChanges={false}
  />

  <svelte:fragment slot="body">
    {#if isLoading}
      <ReconcilingSpinner />
    {:else if error}
      <div
        class="flex flex-col size-full items-center justify-center text-fg-secondary gap-y-1"
      >
        <p class="font-semibold">Unable to read Parquet file</p>
        <p class="text-sm">
          {error.message ?? "Unknown error"}
        </p>
      </div>
    {:else}
      <div class="size-full overflow-hidden border rounded-[2px]">
        <PreviewTable {rows} columnNames={columns} name={fileName} />
      </div>
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
