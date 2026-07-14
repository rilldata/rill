<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { Copy, Columns, PackageOpen, Save } from "lucide-svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import SaveAsComponentDialog from "@rilldata/web-common/features/custom-viz/extract/SaveAsComponentDialog.svelte";
  import { detachComponentRef } from "@rilldata/web-common/features/custom-viz/extract/extract-component";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import { CustomChartComponent } from "./components/charts/custom-chart";
  import { ComponentRefComponent } from "./components/component-ref";
  import type { ComponentWithMetricsView } from "./components/types";
  import ExploreLink from "./explore-link/ExploreLink.svelte";

  export let dropdownOpen = false;
  export let onDelete: () => void;
  export let onDuplicate: () => void;
  // Optional: convert this component's row into a tab group. Only provided for
  // top-level rows (a tab's rows cannot be nested into another tab group).
  export let onConvertToTabGroup: (() => void) | undefined = undefined;
  export let editable = false;
  export let component: BaseCanvasComponent;
  export let navigationEnabled: boolean = true;

  const { customComponents } = featureFlags;

  let saveAsComponentOpen = false;

  $: canSaveAsComponent =
    $customComponents && component instanceof CustomChartComponent;
  $: canDetach =
    $customComponents && component instanceof ComponentRefComponent;

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
          Duplicate
        </DropdownMenu.Item>
        {#if canSaveAsComponent}
          <DropdownMenu.Item onclick={() => (saveAsComponentOpen = true)}>
            <Save size="14px" />
            {m.component_extract_menu_item()}
          </DropdownMenu.Item>
        {/if}
        {#if canDetach}
          <DropdownMenu.Item
            onclick={() => detachComponentRef(component as ComponentRefComponent)}
          >
            <PackageOpen size="14px" />
            {m.component_detach_menu_item()}
          </DropdownMenu.Item>
        {/if}
        {#if onConvertToTabGroup}
          <DropdownMenu.Item onclick={onConvertToTabGroup}>
            <Columns size="14px" />
            Convert row to tab group
          </DropdownMenu.Item>
        {/if}
        {#if showExplore && exploreComponent}
          <DropdownMenu.Separator />
          <ExploreLink component={exploreComponent} mode="dropdown-item" />
        {/if}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          onclick={onDelete}
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

{#if canSaveAsComponent}
  <SaveAsComponentDialog
    bind:open={saveAsComponentOpen}
    component={component as CustomChartComponent}
  />
{/if}
