<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { chartMetadata } from "@rilldata/web-common/features/canvas/components/charts/util";
  import { PlusCircle } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import type { ChartType } from "./components/charts/types";
  import type { CanvasComponentType } from "./components/types";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  import { Tooltip } from "bits-ui";

  type MenuItem = {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  };

  // Function to get a random chart type
  function getRandomChartType(): ChartType {
    const chartTypes = chartMetadata.map((chart) => chart.type);
    const randomIndex = Math.floor(Math.random() * chartTypes.length);
    return chartTypes[randomIndex];
  }

  // Create menu items with a function to get random chart type when clicked
  export const menuItems: MenuItem[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon }, // Default value, will be replaced with random type when clicked
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi_grid", label: "KPI", icon: BigNumberIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];

  export let disabled = false;
  export let componentForm = false;
  export let open = false;
  export let onItemClick: (type: CanvasComponentType) => void;
  export let onMouseEnter: () => void;

  // Wrapper function to handle chart item click with randomization
  function handleChartItemClick() {
    const randomChartType = getRandomChartType();
    onItemClick(randomChartType);
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip.Root>
      <Tooltip.Trigger>
        {#if componentForm}
          <button
            {...builder}
            use:builder.action
            class="pointer-events-auto shadow-sm hover:shadow-md flex bg-white h-[84px] flex-col justify-center gap-2 items-center rounded-md border border-slate-200 w-full"
          >
            <PlusCircle class="w-6 h-6 text-slate-500" />
            <span class="text-sm font-medium text-slate-500">Add widget</span>
          </button>
        {:else}
          <button
            {disabled}
            on:mouseenter={onMouseEnter}
            use:builder.action
            class="pointer-events-auto disabled:pointer-events-none h-7 px-2 grid place-content-center z-50 hover:bg-gray-100 text-slate-500 disabled:opacity-50"
          >
            <PlusCircle size="15px" />
          </button>
        {/if}
      </Tooltip.Trigger>
      <Tooltip.Content side="top" sideOffset={8}>
        <div class="bg-gray-700 text-white rounded p-2 pt-1 pb-1">
          Insert widget
        </div>
      </Tooltip.Content>
    </Tooltip.Root>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align={componentForm ? "center" : "start"}>
    <div class="flex flex-col" role="presentation" on:mouseenter={onMouseEnter}>
      {#each menuItems as { id, label, icon } (id)}
        <DropdownMenu.Item
          on:click={() => {
            open = false;
            if (id === "bar_chart") {
              handleChartItemClick();
            } else {
              onItemClick(id);
            }
          }}
        >
          <div class="flex flex-row gap-x-2">
            <svelte:component this={icon} />
            {label}
          </div>
        </DropdownMenu.Item>
      {/each}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
