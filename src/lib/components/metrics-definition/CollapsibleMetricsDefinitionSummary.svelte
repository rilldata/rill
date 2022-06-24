<script lang="ts">
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { EntityId } from "@reduxjs/toolkit";
  import CollapsibleMetricsDefinitionSummaryNavEntry from "./CollapsibleMetricsDefinitionSummaryNavEntry.svelte";
  import { createReadableStoreWithSelector } from "$lib/redux-store/store-root";
  import { Readable } from "svelte/store";
  import { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { singleMetricsDefSelector } from "$lib/redux-store/metrics-definition-slice";

  export let metricsDefId: EntityId;

  let selectedMetricsDef: Readable<MetricsDefinitionEntity>;
  $: if (metricsDefId) {
    selectedMetricsDef = createReadableStoreWithSelector(
      singleMetricsDefSelector(metricsDefId)
    );
  }

  $: summaryExpanded = $selectedMetricsDef?.summaryExpandedInNav;
  export let indentLevel = 0;
  let containerWidth = 0;
  let contextMenuOpen = false;
  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });

  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }
</script>

<div bind:this={container}>
  <div class="active:cursor-grabbing">
    <CollapsibleMetricsDefinitionSummaryNavEntry {metricsDefId} />
  </div>
  {#if summaryExpanded}
    <div
      class="pt-1 pb-3 pl-accordion"
      transition:slide|local={{ duration: 120 }}
    >
      <div
        class="pl-{indentLevel === 1
          ? '10'
          : '4'} pr-5 pb-2 flex justify-between text-gray-500"
        class:flex-col={containerWidth < 325}
      >
        <em>((summary placeholder))</em>
      </div>
    </div>
  {/if}
</div>
