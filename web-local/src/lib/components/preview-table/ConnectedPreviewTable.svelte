<script lang="ts">
  import {
    useQueryServiceTableRows,
    useQueryServiceTableColumns,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";

  import { PreviewTable } from ".";
  import { runtimeStore } from "../../application-state-stores/application-store";

  export let objectName: string;
  export let limit = 150;

  $: profileColumnsQuery = useQueryServiceTableColumns(
    $runtimeStore?.instanceId,
    objectName,
    {}
  );
  $: profileColumns = $profileColumnsQuery?.data?.profileColumns;

  $: tableQuery = useQueryServiceTableRows(
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
