<script lang="ts">
  import {
    useRuntimeServiceGetNullCount,
    useRuntimeServiceProfileColumns,
    useRuntimeServiceTableCardinality,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import { singleDigitPercentage } from "@rilldata/web-local/lib/util/formatters";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let objectName: string;
  export let columnName: string;

  /**
   * Get the null counts for this profile.
   */
  let nullCountQuery;

  function getColumn(profileColumns, columnName) {
    return profileColumns?.data?.profileColumns?.find(
      (column) => column.name === columnName
    );
  }

  $: profileColumns = useRuntimeServiceProfileColumns(
    $runtimeStore?.instanceId,
    objectName
  );

  $: type = getColumn($profileColumns, columnName)?.type;

  $: if ($runtimeStore?.instanceId)
    nullCountQuery = useRuntimeServiceGetNullCount(
      $runtimeStore?.instanceId,
      objectName,
      columnName
    );

  let nullCount;
  // FIXME: count should not be a string. For now, let's patch it.
  $: nullCount = +$nullCountQuery?.data?.count;

  /**
   * Get the total rows for this profile.
   */
  let totalRowsQuery;
  $: if ($runtimeStore?.instanceId) {
    totalRowsQuery = useRuntimeServiceTableCardinality(
      $runtimeStore?.instanceId,
      objectName
    );
  }
  let totalRows;
  // FIXME: count should not be a string.
  $: totalRows = +$totalRowsQuery?.data?.cardinality;

  $: percentage = nullCount / totalRows;
</script>

{#if $totalRowsQuery?.isSuccess && $nullCountQuery?.isSuccess && totalRows !== undefined && nullCount !== undefined && !isNaN(percentage)}
  <Tooltip location="right" alignment="center" distance={8}>
    <BarAndLabel
      showBackground={nullCount !== 0}
      color={DATA_TYPE_COLORS[type]?.bgClass}
      value={percentage || 0}
    >
      <span class:text-gray-300={nullCount === 0}
        >∅ {singleDigitPercentage(percentage)}</span
      >
    </BarAndLabel>
    <TooltipContent slot="tooltip-content">
      <svelte:fragment slot="title">
        what percentage of values are null?
      </svelte:fragment>
      {#if nullCount > 0}
        {singleDigitPercentage(percentage)} of the values are null
      {:else}
        no null values in this column
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
