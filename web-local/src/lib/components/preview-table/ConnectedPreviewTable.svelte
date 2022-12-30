<script lang="ts">
  import {
    useRuntimeServiceGetTableRows,
    useRuntimeServiceProfileColumns,
  } from "@rilldata/web-common/runtime-client";

  import { PreviewTable } from ".";
  import { runtimeStore } from "../../application-state-stores/application-store";

  export let objectName: string;
  export let limit = 150;

  $: profileColumnsQuery = useRuntimeServiceProfileColumns(
    $runtimeStore?.instanceId,
    objectName,
    {}
  );
  $: profileColumns = $profileColumnsQuery?.data?.profileColumns;

  $: tableQuery = useRuntimeServiceGetTableRows(
    $runtimeStore?.instanceId,
    objectName,
    { limit }
  );

  $: rows = $tableQuery?.data?.data;
</script>

{#if rows && profileColumns}
  <PreviewTable {rows} columnNames={profileColumns} />
{/if}
