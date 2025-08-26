<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { BarChart, LineChart } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

  type MARK_TYPES = "bar" | "line";

  export let selectedMark: MARK_TYPES = "bar";
  export let onClick: (mark: MARK_TYPES) => void;

  const markOptions: {
    mark: MARK_TYPES;
    icon: ComponentType<SvelteComponent>;
  }[] = [
    { mark: "bar", icon: BarChart },
    { mark: "line", icon: LineChart },
  ];
</script>

<div class="flex items-center gap-2 justify-between mt-2 px-2">
  <InputLabel
    small
    capitalize={false}
    label="Mark type"
    id="mark-type-toggle"
  />
  <div class="radio relative">
    {#each markOptions as { mark, icon: Icon } (mark)}
      <Tooltip activeDelay={700} distance={8}>
        <button
          aria-label="Switch to {mark === 'bar' ? 'bar' : 'line'} editor"
          id="{mark}-toggle"
          class="size-[24px] z-10 hover:brightness-75"
          on:click={() => {
            selectedMark = mark;
            onClick(mark);
          }}
        >
          <Icon
            size="15px"
            color={mark === selectedMark ? "#374151" : "#D1D5DB"}
          />
        </button>
        <TooltipContent slot="tooltip-content">
          {mark === "bar" ? "Bar chart" : "Line chart"}
        </TooltipContent>
      </Tooltip>
    {/each}
    <span
      style:left={selectedMark === "bar" ? "2px" : "24px"}
      class="toggle size-[24px] pointer-events-none absolute rounded-[4px] z-0 transition-[left]"
    />
  </div>
</div>

<style lang="postcss">
  button {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .toggle {
    @apply bg-surface outline outline-slate-200 outline-[1px];
  }

  .radio {
    @apply h-fit bg-slate-100 p-[2px] rounded-[6px] flex;
  }
</style>
