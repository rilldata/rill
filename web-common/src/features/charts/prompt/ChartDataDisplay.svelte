<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import { useVariableInputParams } from "@rilldata/web-common/features/custom-dashboards/variables-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import {
    createQueryServiceResolveComponent,
    V1ComponentSpecResolverProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  export let chartName: string;
  export let tablePercentage = 0.45;
  export let containerHeight: number;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;
  export let input: V1ComponentVariable[] | undefined;

  const dashboardName = getContext("rill::custom-dashboard:name") as string;

  $: ({ instanceId } = $runtime);

  $: tableHeight = tablePercentage * containerHeight;

  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: chartDataQuery = resolverProperties
    ? createQueryServiceResolveComponent(instanceId, chartName, {
        args: $inputVariableParams,
      })
    : null;

  $: chartData = $chartDataQuery?.data?.data;
  $: isFetching = $chartDataQuery?.isFetching ?? false;
  $: errorMessage = $chartDataQuery?.error?.response?.data?.message;

  $: console.log($chartDataQuery, resolverProperties, errorMessage, chartData);
</script>

{#if resolverProperties}
  <div
    class="size-full h-48 bg-gray-100 border-t relative flex-none flex-shrink-0"
    style:height="{tablePercentage * 100}%"
  >
    <Resizer
      direction="NS"
      dimension={tableHeight}
      min={100}
      max={0.65 * containerHeight}
      onUpdate={(height) => (tablePercentage = height / containerHeight)}
    />

    {#if isFetching}
      <div class="flex flex-col gap-y-2 size-full justify-center items-center">
        <Spinner size="2em" status={EntityStatus.Running} />
        <div>Loading chart data</div>
      </div>
    {:else if errorMessage}
      <div class="size-full flex flex-col items-center justify-center">
        <div class="flex items-center gap-x-2 text-lg">
          <CancelCircle /> Error fetching data from SQL
        </div>
        <div class="text-sm">{errorMessage}</div>
      </div>
    {:else if chartData}
      <PreviewTable
        rows={chartData}
        name={chartName}
        columnNames={Object.keys(chartData[0]).map((key) => ({
          type: "VARCHAR",
          name: key,
        }))}
      />
    {:else}
      <p class="text-lg size-full grid place-content-center">
        Update or add SQL to view data
      </p>
    {/if}
  </div>
{/if}
