<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { Copy, Columns, Download } from "lucide-svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "./components/types";
  import ExploreLink from "./explore-link/ExploreLink.svelte";

  export let dropdownOpen = false;
  export let onDelete: () => void;
  export let onDuplicate: () => void;
  export let onDownloadPng: () => void;
  // Optional: convert this component's row into a tab group. Only provided for
  // top-level rows (a tab's rows cannot be nested into another tab group).
  export let onConvertToTabGroup: (() => void) | undefined = undefined;
  export let editable = false;
  export let component: BaseCanvasComponent;
  export let navigationEnabled: boolean = true;

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
    "combo_chart",
    "custom_chart",
  ] as const;

  $: showExplore =
    navigationEnabled &&
    EXPLORE_SUPPORTED_TYPES.includes(component.type as any);
  $: exploreComponent = showExplore
    ? (component as BaseCanvasComponent<ComponentWithMetricsView>)
    : null;
</script>

<div
  class:!flex={dropdownOpen}
  class="group-hover:flex p-0 overflow-hidden bg-surface-card gap-x-1 items-center justify-center hidden toolbar top-0 right-0 shadow-sm z-[1000] absolute w-fit border-l border-b pointer-events-auto rounded-bl-sm rounded-tr-sm"
>
  {#if editable}
    <!-- Editable mode: Show dropdown with explore option -->
    <DropdownMenu.Root bind:open={dropdownOpen}>
      <DropdownMenu.Trigger
        class="size-7 grid place-content-center bg-surface-card hover:brightness-[85%] active:brightness-75"
        aria-label={m.canvas_component_menu_aria()}
      >
        <ThreeDot size="16px" />
      </DropdownMenu.Trigger>

      <DropdownMenu.Content
        align="end"
        sideOffset={8}
        alignOffset={-4}
        class="w-40"
      >
        <DropdownMenu.Item onclick={onDuplicate}>
          <Copy size="14px" />
          {m.canvas_component_duplicate()}
        </DropdownMenu.Item>
        {#if onConvertToTabGroup}
          <DropdownMenu.Item onclick={onConvertToTabGroup}>
            <Columns size="14px" />
            {m.canvas_component_convert_to_tab_group()}
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Separator />
        {#if showExplore && exploreComponent}
          <ExploreLink component={exploreComponent} mode="dropdown-item" />
        {/if}
        <DropdownMenu.Item onclick={onDownloadPng}>
          <Download size="14px" />
          {m.dashboard_download_as_png()}
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          onclick={onDelete}
          class="text-red-600 data-[highlighted]:text-red-600"
        >
          <Trash size="14px" />
          {m.canvas_delete()}
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {:else}
    <!-- Non-editable mode: one-click explore jump plus a menu for the rest -->
    {#if showExplore && exploreComponent}
      <ExploreLink component={exploreComponent} mode="icon-button" />
    {/if}
    <DropdownMenu.Root bind:open={dropdownOpen}>
      <DropdownMenu.Trigger
        class="size-7 grid place-content-center bg-surface-card hover:brightness-[85%] active:brightness-75"
        aria-label={m.canvas_component_menu_aria()}
      >
        <ThreeDot size="16px" />
      </DropdownMenu.Trigger>

      <DropdownMenu.Content
        align="end"
        sideOffset={8}
        alignOffset={-4}
        class="w-44"
      >
        <DropdownMenu.Item onclick={onDownloadPng}>
          <Download size="14px" />
          {m.dashboard_download_as_png()}
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>
