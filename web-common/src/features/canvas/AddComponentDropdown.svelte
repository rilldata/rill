<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { CHART_TYPES } from "@rilldata/web-common/features/components/charts/config";
  import { Plus, PlusCircle } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import type { ChartType } from "../components/charts/types";
  import type { CanvasComponentType } from "./components/types";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import LeaderboardIcon from "./icons/LeaderboardIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  type MenuItem = {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  };

  // Function to get a random chart type
  function getRandomChartType(): ChartType {
    const chartTypes = CHART_TYPES.filter(
      (t) => t !== "stacked_bar_normalized",
    );
    const randomIndex = Math.floor(Math.random() * chartTypes.length);
    return chartTypes[randomIndex] as ChartType;
  }

  // Create menu items with a function to get random chart type when clicked
  export const menuItems: MenuItem[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon }, // Default value, will be replaced with random type when clicked
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi_grid", label: "KPI", icon: BigNumberIcon },
    { id: "leaderboard", label: "Leaderboard", icon: LeaderboardIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];

  export let disabled = false;
  export let componentForm = false;
  export let floatingForm = false;
  export let open = false;
  export let rowIndex: number | undefined = undefined;
  export let columnIndex: number | undefined = undefined;
  export let onItemClick: (type: CanvasComponentType) => void;
  export let onMouseEnter: () => void = () => {};
  export let onOpenChange: (isOpen: boolean) => void = () => {};

  // Wrapper function to handle chart item click with randomization
  function handleChartItemClick() {
    const randomChartType = getRandomChartType();
    onItemClick(randomChartType);
  }

  function getAriaLabel(row: number | undefined, column: number | undefined) {
    return `Insert widget${row !== undefined ? ` in row ${row + 1}` : ""}${
      column !== undefined ? ` at column ${column + 1}` : ""
    }`;
  }
</script>

<DropdownMenu.Root bind:open {onOpenChange}>
  <DropdownMenu.Trigger asChild let:builder>
    {#if componentForm}
      <button
        {...builder}
        use:builder.action
        class="pointer-events-auto shadow-sm hover:shadow-md flex bg-surface h-[84px] flex-col justify-center gap-2 items-center rounded-md border border-slate-200 w-full"
      >
        <PlusCircle class="w-6 h-6 text-slate-500" />
        <span class="text-sm font-medium text-slate-500">Add widget</span>
      </button>
    {:else if floatingForm}
      <button
        {...builder}
        use:builder.action
        class:pr-3.5={open}
        aria-label="Add widget"
        class="shadow-lg flex group hover:rounded-3xl w-fit gap-x-1 p-2 hover:pr-3.5 absolute bottom-3 right-3 items-center justify-center z-50 rounded-full bg-primary-600 text-white hover:bg-primary-500"
      >
        <Plus size="20px" />

        <span
          class:not-sr-only={open}
          class="sr-only group-hover:not-sr-only font-semibold w-fit"
        >
          Add widget
        </span>
      </button>
    {:else}
      <button
        {disabled}
        use:builder.action
        {...builder}
        aria-label={getAriaLabel(rowIndex, columnIndex)}
        title="Insert widget"
        class:bg-gray-50={open}
        class="pointer-events-auto bg-surface active:bg-gray-100 disabled:pointer-events-none h-7 px-2 grid place-content-center z-50 hover:bg-gray-50 text-slate-500 disabled:opacity-50"
        on:mouseenter={onMouseEnter}
      >
        <PlusCircle size="15px" />
      </button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align={componentForm || floatingForm ? "center" : "start"}
  >
    <div class="flex flex-col" role="presentation" on:mouseenter={onMouseEnter}>
      {#each menuItems as { id, label, icon } (id)}
        <DropdownMenu.Item
          class="flex flex-row gap-x-2"
          on:click={() => {
            if (id === "bar_chart") {
              handleChartItemClick();
            } else {
              onItemClick(id);
            }
          }}
        >
          <svelte:component this={icon} color="var(--color-gray-600)" />
          {label}
        </DropdownMenu.Item>
      {/each}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
