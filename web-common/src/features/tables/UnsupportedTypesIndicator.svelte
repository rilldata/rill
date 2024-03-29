<script lang="ts">
  import WarningIcon from "../../components/icons/WarningIcon.svelte";
  import Tooltip from "../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../components/tooltip/TooltipContent.svelte";
  import { createQueryServiceTableColumns } from "../../runtime-client";

  export let instanceId: string;
  export let connector: string;
  export let tableName: string;

  $: tableColumns = createQueryServiceTableColumns(instanceId, tableName, {
    connector: connector,
  });

  $: unsupportedColumnsMap = $tableColumns.data?.unsupportedColumns;
  $: unsupportedColumns = unsupportedColumnsMap
    ? Object.entries(unsupportedColumnsMap).map(([column, type]) => ({
        column,
        type,
      }))
    : [];
</script>

<Tooltip distance={8} alignment="start">
  <WarningIcon />
  <TooltipContent slot="tooltip-content">
    This table contains columns with unsupported data types:

    <ul class="list-disc pl-4 mt-1">
      {#each unsupportedColumns as { column, type } (column)}
        <li>{column}: {type}</li>
      {/each}
    </ul>
  </TooltipContent>
</Tooltip>
