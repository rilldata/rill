<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import {
    useRuntimeServiceGetTableCardinality,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";

  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import { derived } from "svelte/store";
  import { slide } from "svelte/transition";

  export let source: V1Source;
  $: embeds = source?.embeds;
  $: modelsAndRowCounts = derived(
    embeds.map((modelName) => {
      return derived(
        useRuntimeServiceGetTableCardinality(
          $runtimeStore?.instanceId,
          modelName
        ),

        (totalRows) => {
          return {
            modelName,
            totalRows: +totalRows?.data?.cardinality,
          };
        }
      );
    }),
    ($row) => $row
  );

  let showModelReferences = true;
</script>

<div class="p-4">
  <div class="pb-1">
    <CollapsibleSectionTitle
      bind:active={showModelReferences}
      tooltipText="referenced models"
      >Used in these models</CollapsibleSectionTitle
    >
  </div>
  {#if showModelReferences}
    <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
      {#each $modelsAndRowCounts as { modelName, totalRows }}
        <Tooltip>
          <a
            href="/model/{modelName}"
            class="grid justify-between gap-x-2 py-1 ui-copy-muted"
            style:grid-template-columns="auto max-content"
          >
            <div
              class="text-ellipsis overflow-hidden whitespace-nowrap flex items-center gap-x-2"
            >
              <div class=" text-ellipsis overflow-hidden whitespace-nowrap">
                {modelName}
              </div>
            </div>

            <div class="text-gray-500">
              {#if totalRows && !isNaN(totalRows)}
                {`${formatCompactInteger(totalRows)} rows`}
              {/if}
            </div>
          </a>
          <TooltipContent slot="tooltip-content">
            <TooltipTitle
              ><svelte:fragment slot="name">{modelName}</svelte:fragment>
              <svelte:fragment slot="description">model</svelte:fragment
              ></TooltipTitle
            >
            <TooltipShortcutContainer>
              <div>Open in workspace</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      {/each}
    </div>
  {/if}
</div>
