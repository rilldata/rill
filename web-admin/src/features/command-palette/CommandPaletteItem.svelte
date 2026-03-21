<script lang="ts">
  import { Folders } from "lucide-svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import type { SearchableItem } from "./types";

  export let item: SearchableItem;

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

<div class="flex items-center gap-2.5 w-full">
  <div class="flex-none w-4 h-4 text-gray-400">
    {#if item.type === "project"}
      <Folders size={16} />
    {:else if IconComponent}
      <svelte:component this={IconComponent} size="16px" />
    {/if}
  </div>
  <div class="flex flex-col min-w-0">
    <span class="text-sm text-gray-200 truncate">{item.name}</span>
    {#if showBreadcrumb}
      <span class="text-xs text-gray-500 truncate">{item.projectName}</span>
    {/if}
  </div>
</div>
