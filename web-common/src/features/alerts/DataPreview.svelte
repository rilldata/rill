<script lang="ts">
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { get } from "svelte/store";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import PreviewTable from "../../components/preview-table/PreviewTable.svelte";
  import type { VirtualizedTableColumns } from "../../components/virtualized-table/types";
  import {
    createQueryServiceMetricsViewAggregation,
    V1Expression,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { getStateManagers } from "../dashboards/state-managers/state-managers";
  import { getLabelForFieldName } from "./utils";

  export let metricsView: string;
  export let measure: string;
  export let dimension: string;
  export let filter: V1Expression;
  export let criteria: V1Expression | undefined = undefined;

  const ctx = getStateManagers();
  const timeControls = get(ctx.selectors.timeRangeSelectors.timeControlsState);

  $: aggregation = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    metricsView,
    {
      measures: [{ name: measure }],
      dimensions: dimension ? [{ name: dimension }] : [],
      where: sanitiseExpression(filter),
      having: sanitiseExpression(criteria),
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    },
    {
      query: {
        enabled: !!measure,
        select: (data) => {
          const rows = data.data;
          const schema = data.schema?.fields?.map((field) => {
            return {
              name: field.name,
              type: field.type?.code,
              label: getLabelForFieldName(ctx, field.name as string),
            };
          }) as VirtualizedTableColumns[];
          return { rows, schema };
        },
      },
    },
  );

  // TODO: throttle fetches
</script>

{#if !$aggregation.data}
  <div class="pt-5 pb-10 flex flex-col justify-center items-center gap-1">
    <TableIcon size="32px" className="text-slate-300" />
    <div class="flex flex-col justify-center items-center text-sm">
      <div class="text-gray-600 font-semibold">No data to preview</div>
      <div class="text-gray-500 font-normal">
        To see preview, select measures above.
      </div>
    </div>
  </div>
{:else}
  <div class="max-h-64 overflow-auto">
    <PreviewTable
      rows={$aggregation.data.rows}
      columnNames={$aggregation.data.schema}
    />
  </div>
{/if}
