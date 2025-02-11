<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let fields: {
    id: string;
    Icon: ComponentType<SvelteComponent>;
    tooltip?: string;
  }[];
  export let selected: string[] = [];
  export let onChange: (values: string[]) => void = () => {};
  export let small = false;
  export let expand = false;

  function toggleSelection(id: string) {
    const newSelected = selected.includes(id)
      ? selected.filter((item) => item !== id)
      : [...selected, id];
    onChange(newSelected);
  }
</script>

<div class:small class:expand class="option-wrapper">
  {#each fields as { id, Icon, tooltip } (id)}
    <Tooltip distance={4} location="top" alignment="start">
      <TooltipContent slot="tooltip-content" maxWidth="280px">
        {tooltip}
      </TooltipContent>
      <button
        on:click={() => toggleSelection(id)}
        class="-ml-[1px] first-of-type:-ml-0 px-2 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
        class:selected={selected.includes(id)}
      >
        <Icon size={small ? "14px" : "16px"} />
      </button>
    </Tooltip>
  {/each}
</div>

<style lang="postcss">
  button {
    @apply flex justify-center items-center;
    @apply bg-primary-50 capitalize;
  }

  button:hover:not(.selected) {
    @apply bg-primary-100;
  }

  .option-wrapper {
    @apply text-primary-500 text-sm w-fit mb-1;
    @apply flex gap-x-0.5 h-6 rounded-[2px];
  }

  .option-wrapper.small {
    @apply h-6 text-xs;
  }

  .expand {
    @apply w-full;
  }
  .expand button {
    @apply flex-1;
  }

  .option-wrapper .selected {
    @apply bg-primary-200 z-50;
  }
</style>
