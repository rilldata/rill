<script lang="ts">
  import { Folders } from "lucide-svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import type { SearchableItem } from "./types";

  export let item: SearchableItem;
  export let loading = false;

  const iconComponents = {
    project: null,
    explore: ExploreIcon,
    canvas: CanvasIcon,
    report: ReportIcon,
    alert: AlertIcon,
  };

  $: IconComponent = iconComponents[item.type];
  $: showBreadcrumb = item.type !== "project";
</script>

<div class="palette-item-content">
  <span class="palette-item-icon">
    {#if loading}
      <LoadingSpinner size="16px" />
    {:else if item.type === "project"}
      <Folders size={16} />
    {:else if IconComponent}
      <svelte:component this={IconComponent} size="16px" />
    {/if}
  </span>
  <span class="palette-item-name">{item.name}</span>
  {#if showBreadcrumb}
    <span class="palette-item-meta">{item.projectName}</span>
  {/if}
</div>

<style>
  .palette-item-content {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    min-width: 0;
  }

  .palette-item-icon {
    flex-shrink: 0;
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--icon-muted);
  }

  .palette-item-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .palette-item-meta {
    margin-left: auto;
    color: var(--fg-muted);
    font-size: 13px;
    flex-shrink: 0;
  }
</style>
