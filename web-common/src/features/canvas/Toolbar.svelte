<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { Copy } from "lucide-svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "./components/types";
  import ExploreLink from "./explore-link/ExploreLink.svelte";

  export let dropdownOpen = false;
  export let onDelete: () => void;
  export let onDuplicate: () => void;
  export let editable = false;
  export let component: BaseCanvasComponent;

  // Component types that support link to explore functionality
  const EXPLORE_SUPPORTED_TYPES = [
    "kpi_grid",
    "leaderboard",
    "table",
    "pivot",
    "bar_chart",
    "line_chart",
    "area_chart",
    "stacked_bar",
    "stacked_bar_normalized",
    "donut_chart",
    "pie_chart",
    "heatmap",
  ] as const;

  $: showExplore = EXPLORE_SUPPORTED_TYPES.includes(component.type as any);
  $: exploreComponent = showExplore
    ? (component as BaseCanvasComponent<ComponentWithMetricsView>)
    : null;
</script>

<div
  class:!flex={dropdownOpen}
  class="group-hover:flex p-0 overflow-hidden bg-slate-50 gap-x-1 items-center justify-center hidden toolbar top-0 right-0 shadow-sm z-[1000] absolute w-fit border-l border-b pointer-events-auto rounded-bl-sm rounded-tr-sm"
>
  {#if editable}
    <!-- Editable mode: Show dropdown with explore option -->
    <DropdownMenu.Root bind:open={dropdownOpen}>
      <DropdownMenu.Trigger
        class="size-7 grid place-content-center hover:bg-slate-100 active:bg-slate-200"
      >
        <ThreeDot size="16px" />
      </DropdownMenu.Trigger>

      <DropdownMenu.Content
        align="end"
        sideOffset={8}
        alignOffset={-4}
        class="w-40"
      >
        <DropdownMenu.Item on:click={onDuplicate}>
          <Copy size="14px" />
          Duplicate
        </DropdownMenu.Item>
        {#if showExplore && exploreComponent}
          <DropdownMenu.Separator />
          <ExploreLink component={exploreComponent} mode="dropdown-item" />
        {/if}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          on:click={onDelete}
          class="text-red-600 data-[highlighted]:text-red-600"
        >
          <Trash size="14px" />
          Delete
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {:else if showExplore && exploreComponent}
    <!-- Non-editable mode: Show explore icon button -->
    <ExploreLink component={exploreComponent} mode="icon-button" />
  {/if}
</div>
