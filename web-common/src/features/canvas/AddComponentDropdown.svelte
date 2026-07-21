<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    CHART_CONFIG,
    VISIBLE_CHART_TYPES,
  } from "@rilldata/web-common/features/components/charts/config";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Layers, Plus, PlusCircle } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import type { ChartType } from "../components/charts/types";
  import type { CanvasComponentType } from "./components/types";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import LeaderboardIcon from "./icons/LeaderboardIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  type MainMenuItem = {
    id: Exclude<CanvasComponentType, ChartType> | "chart_submenu";
    label: string;
    icon: ComponentType<SvelteComponent>;
  };

  // Create menu items with a function to get random chart type when clicked
  export const menuItems: MainMenuItem[] = [
    { id: "chart_submenu", label: m.canvas_chart(), icon: ChartIcon },
    { id: "table", label: m.canvas_table(), icon: TableIcon },
    { id: "markdown", label: m.canvas_text_markdown(), icon: TextIcon },
    { id: "kpi_grid", label: m.canvas_kpi(), icon: BigNumberIcon },
    { id: "leaderboard", label: m.canvas_leaderboard(), icon: LeaderboardIcon },
    { id: "image", label: m.canvas_image(), icon: ChartIcon },
  ];

  export let disabled = false;
  export let componentForm = false;
  export let floatingForm = false;
  // Label shown on the large (componentForm) add button.
  export let label = m.canvas_add_widget();
  export let open = false;
  export let rowIndex: number | undefined = undefined;
  export let columnIndex: number | undefined = undefined;
  export let onItemClick: (type: CanvasComponentType) => void;
  export let onMouseEnter: () => void = () => {};
  export let onOpenChange: (isOpen: boolean) => void = () => {};
  // When provided, the menu offers "Tab group" as a final item. Only passed at the
  // top level (tab groups cannot be nested inside a tab or a column).
  export let onAddTabGroup: (() => void) | undefined = undefined;

  const { customCharts } = featureFlags;

  const ADD_DROPDOWN_CHART_TYPES = VISIBLE_CHART_TYPES.filter((type) => {
    return type !== "stacked_bar" && type !== "stacked_bar_normalized";
  });

  function getAriaLabel(row: number | undefined, column: number | undefined) {
    if (row !== undefined && column !== undefined) {
      return m.canvas_insert_widget_at({
        row: String(row + 1),
        col: String(column + 1),
      });
    }
    if (row !== undefined) {
      return m.canvas_insert_widget_at_row({ row: String(row + 1) });
    }
    if (column !== undefined) {
      return m.canvas_insert_widget_at_col({ col: String(column + 1) });
    }
    return m.canvas_insert_widget();
  }
</script>

<DropdownMenu.Root bind:open {onOpenChange}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      {#if componentForm}
        <button
          {...props}
          {disabled}
          class="pointer-events-auto shadow-sm hover:shadow-md flex bg-surface-subtle h-[84px] flex-col justify-center gap-2 items-center rounded-md border border-gray-200 w-full"
        >
          <PlusCircle class="w-6 h-6 text-fg-secondary" />
          <span class="text-sm font-medium text-fg-secondary">{label}</span>
        </button>
      {:else if floatingForm}
        <button
          {...props}
          {disabled}
          class:pr-3.5={open}
          aria-label={m.canvas_add_widget()}
          class="shadow-lg flex group hover:rounded-3xl w-fit gap-x-1 p-2 hover:pr-3.5 absolute bottom-3 right-3 items-center justify-center z-50 rounded-full bg-primary-600 text-white hover:bg-primary-500"
        >
          <Plus size="20px" />

          <span
            class:not-sr-only={open}
            class="sr-only group-hover:not-sr-only font-semibold w-fit"
          >
            {m.canvas_add_widget()}
          </span>
        </button>
      {:else}
        <button
          {...props}
          {disabled}
          aria-label={getAriaLabel(rowIndex, columnIndex)}
          title={m.canvas_insert_widget()}
          class:bg-surface-background={open}
          class="pointer-events-auto bg-surface-subtle active:bg-gray-100 disabled:pointer-events-none h-7 px-2 grid place-content-center z-50 hover:bg-surface-background text-fg-secondary disabled:opacity-50"
          onmouseenter={onMouseEnter}
        >
          <PlusCircle size="15px" />
        </button>
      {/if}
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align={componentForm || floatingForm ? "center" : "start"}
  >
    <div class="flex flex-col" role="presentation" onmouseenter={onMouseEnter}>
      {#each menuItems as { id, label, icon } (id)}
        {#if id === "chart_submenu"}
          <DropdownMenu.Sub>
            <DropdownMenu.SubTrigger class="flex flex-row gap-x-2">
              <svelte:component this={icon} />
              {label}
            </DropdownMenu.SubTrigger>
            <DropdownMenu.SubContent class="min-w-[160px]">
              {#each ADD_DROPDOWN_CHART_TYPES as chartType (chartType)}
                <DropdownMenu.Item
                  class="flex flex-row gap-x-2 text-fg-primary"
                  onclick={() => onItemClick(chartType)}
                >
                  <svelte:component
                    this={CHART_CONFIG[chartType].icon}
                    primaryColor="#111827"
                    secondaryColor="#9ca3af"
                  />
                  {CHART_CONFIG[chartType].title}
                </DropdownMenu.Item>
              {/each}
              {#if $customCharts}
                <DropdownMenu.Separator />
                <DropdownMenu.Item
                  class="flex flex-row gap-x-2 text-fg-primary"
                  onclick={() => onItemClick("custom_chart")}
                >
                  <ChartIcon />
                  {m.canvas_custom_chart()}
                </DropdownMenu.Item>
              {/if}
            </DropdownMenu.SubContent>
          </DropdownMenu.Sub>
        {:else}
          <DropdownMenu.Item
            class="flex flex-row gap-x-2 text-fg-primary"
            onclick={() => onItemClick(id)}
          >
            <svelte:component this={icon} />
            {label}
          </DropdownMenu.Item>
        {/if}
      {/each}

      {#if onAddTabGroup}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex flex-row gap-x-2 text-fg-primary"
          onclick={() => onAddTabGroup?.()}
        >
          <Layers size="16px" />
          {m.canvas_tab_group()}
        </DropdownMenu.Item>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
