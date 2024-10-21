<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import { useVariableInputParams } from "@rilldata/web-common/features/canvas/variables-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    createQueryServiceResolveComponent,
    type V1ComponentSpecResolverProperties,
    type V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  export let componentName: string;
  export let tablePercentage = 0.45;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;
  export let input: V1ComponentVariable[] | undefined;

  const canvasName = getContext("rill::canvas:name") as string;

  $: ({ instanceId } = $runtime);

  $: inputVariableParams = useVariableInputParams(canvasName, input);

  $: componentDataQuery = resolverProperties
    ? createQueryServiceResolveComponent(instanceId, componentName, {
        args: $inputVariableParams,
      })
    : null;

  $: componentData = $componentDataQuery?.data?.data;
  $: isFetching = $componentDataQuery?.isFetching ?? false;
  $: errorMessage = $componentDataQuery?.error?.response?.data?.message;
</script>

{#if resolverProperties}
  <div
    class="w-full h-48 relative flex-none border rounded-[2px] overflow-hidden"
    style:height="{tablePercentage * 100}%"
  >
    {#if isFetching}
      <div class="flex flex-col gap-y-2 size-full justify-center items-center">
        <Spinner size="2em" status={EntityStatus.Running} />
        <div>Loading component data</div>
      </div>
    {:else if errorMessage}
      <div class="size-full flex flex-col items-center justify-center">
        <div class="flex items-center gap-x-2 text-lg">
          <CancelCircle /> Error fetching data from SQL
        </div>
        <div class="text-sm">{errorMessage}</div>
      </div>
    {:else if componentData}
      <PreviewTable
        rows={componentData}
        name={componentName}
        columnNames={Object.keys(componentData[0]).map((key) => ({
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
