<script lang="ts">
  import MetricsViewIcon from "../../../../components/icons/MetricsViewIcon.svelte";
  import Tooltip from "../../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../../components/tooltip/TooltipContent.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../../../entity-management/resource-selectors";
  import MetricsCatalogDialog from "./MetricsCatalogDialog.svelte";

  $: ({ instanceId } = $runtime);

  $: metricsViewsQuery = useFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
  );

  $: metricsViews = $metricsViewsQuery?.data ?? [];
  $: metricsCount = metricsViews.length;

  let dialogOpen = false;

  function openCatalog() {
    dialogOpen = true;
  }

  function closeCatalog() {
    dialogOpen = false;
  }
</script>

<Tooltip distance={8}>
  <button class="metrics-pill" on:click={openCatalog}>
    <MetricsViewIcon size="14px" color="#6366f1" />
    <span class="metrics-text"
      >{metricsCount}
      {metricsCount === 1 ? "metrics view" : "metrics views"}</span
    >
  </button>
  <TooltipContent slot="tooltip-content"
    >Browse your available metrics</TooltipContent
  >
</Tooltip>

<MetricsCatalogDialog {metricsViews} open={dialogOpen} onClose={closeCatalog} />

<style>
  .metrics-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.25rem 0.625rem;
    background: #f3f4f6;
    border: 1px solid #e5e7eb;
    border-radius: 9999px;
    font-size: 0.75rem;
    color: #374151;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
    width: fit-content;
  }

  .metrics-pill:hover {
    background: #e5e7eb;
    border-color: #d1d5db;
  }

  .metrics-text {
    line-height: 1;
  }
</style>
