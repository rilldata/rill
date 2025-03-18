<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { chartMetadata } from "@rilldata/web-common/features/canvas/components/charts/util";
  import { Plus, PlusCircle } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import type { ChartType } from "./components/charts/types";
  import type { CanvasComponentType } from "./components/types";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { hoveredDivider } from "./stores/ui-stores";

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
  export let floatingForm = false;
  export let open = false;
  export let dividerId: string | null = null;

  export let onItemClick: (type: CanvasComponentType) => void;
  export let onMouseEnter: () => void = () => {};

  function onOpenChange(isOpen: boolean) {
    if (!dividerId) return;
    if (isOpen) {
      console.log("claiming active");
      hoveredDivider.setActive(dividerId, true);
    } else {
      console.log("resetttinggg");
      hoveredDivider.reset(0);
    }
  }

  // Wrapper function to handle chart item click with randomization
  function handleChartItemClick() {
    const randomChartType = getRandomChartType();
    onItemClick(randomChartType);
  }
</script>

<DropdownMenu.Root bind:open {onOpenChange}>
  <DropdownMenu.Trigger asChild let:builder>
    {#if componentForm}
      <button
        {...builder}
        use:builder.action
        class="pointer-events-auto shadow-sm hover:shadow-md flex bg-white h-[84px] flex-col justify-center gap-2 items-center rounded-md border border-slate-200 w-full"
      >
        <PlusCircle class="w-6 h-6 text-slate-500" />
        <span class="text-sm font-medium text-slate-500">Add widget</span>
      </button>
    {:else if floatingForm}
      <button
        {...builder}
        use:builder.action
        class="shadow-lg flex group hover:rounded-3xl w-fit p-2 absolute bottom-3 right-3 items-center justify-center z-50 rounded-full bg-primary-600 text-white hover:bg-primary-500"
      >
        <Plus size="20px" />
        <span
          class:w-[80px]={open}
          class:opacity-100={open}
          class="w-0 overflow-hidden text-clip line-clamp-1 font-semibold opacity-0 group-hover:opacity-100 group-hover:w-[80px] transition-[width]"
        >
          Add widget
        </span>
      </button>
    {:else}
      <Tooltip distance={8} location="top" suppress={open}>
        <button
          {disabled}
          {...builder}
          use:builder.action
          on:mouseenter={onMouseEnter}
          class="pointer-events-auto disabled:pointer-events-none h-7 px-2 grid place-content-center z-50 hover:bg-gray-100 text-slate-500 disabled:opacity-50"
        >
          <PlusCircle size="15px" />
        </button>

        <TooltipContent slot="tooltip-content">Insert widget</TooltipContent>
      </Tooltip>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align={componentForm || floatingForm ? "center" : "start"}
  >
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
