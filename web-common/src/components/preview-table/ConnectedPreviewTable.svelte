<script lang="ts">
  import {
    useRuntimeServiceGetTableRows,
    useRuntimeServiceProfileColumns,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { onMount } from "svelte";
  import { PreviewTable } from ".";

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

  /** We will set the overscan amounts to 0 for initial render;
   * in practice, this will shave off around 200ms from the initial render.
   * Then, after 1 second, we will set the overscan amounts to 40 and 10,
   * which wil then cause the table to render with the overscan amounts.
   */
  let rowOverscanAmount = 0;
  let columnOverscanAmount = 0;
  onMount(() => {
    setTimeout(() => {
      rowOverscanAmount = 40;
      columnOverscanAmount = 10;
    }, 1000);
  });
</script>

{#if rows && profileColumns}
  <PreviewTable
    {rows}
    columnNames={profileColumns}
    {rowOverscanAmount}
    {columnOverscanAmount}
  />
{/if}
