<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { Copy, Columns, Download, FileSpreadsheet } from "lucide-svelte";
  import { buildExportFilename } from "@rilldata/web-common/features/exports/filename";
  import { componentCsv } from "@rilldata/web-common/features/exports/csv";
  import { downloadBlob } from "@rilldata/web-common/features/exports/xlsx/client";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { get } from "svelte/store";
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

  // The data the component shows, as CSV with values unformatted;
  // undefined when it has none to download. Worked out only while the menu is open.
  $: exportPayload = component.exportPayload;
  $: csv = dropdownOpen
    ? componentCsv(
        $exportPayload,
        get(component.timeAndFilterStore)?.timeGrain,
        get(component.parent.timeManager.state.timeZoneStore) || "UTC",
      )
    : undefined;

  function downloadCsv() {
    if (csv === undefined) return;
    const { title } = get(component.specStore) as { title?: unknown };
    downloadBlob(
      new Blob([csv], { type: "text/csv" }),
      buildExportFilename(
        (typeof title === "string" && title) ||
          get(component.parent.titleStore) ||
          component.id,
        "csv",
      ),
    );
  }
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
        {#if csv !== undefined}
          <DropdownMenu.Item onclick={downloadCsv}>
            <FileSpreadsheet size="14px" />
            {m.dashboard_download_as_csv()}
          </DropdownMenu.Item>
        {/if}
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
    <!-- Non-editable mode: the menu sits inward of the explore jump so the
         corner slot users already aim for keeps taking them to explore. -->
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
        {#if csv !== undefined}
          <DropdownMenu.Item onclick={downloadCsv}>
            <FileSpreadsheet size="14px" />
            {m.dashboard_download_as_csv()}
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
    {#if showExplore && exploreComponent}
      <ExploreLink component={exploreComponent} mode="icon-button" />
    {/if}
  {/if}
</div>
