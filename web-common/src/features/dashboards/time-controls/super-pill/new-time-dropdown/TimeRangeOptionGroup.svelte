<script lang="ts">
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";
  import TimeRangeMenuItem from "../components/TimeRangeMenuItem.svelte";

  export let filter = "";
  export let options: RillTime[];
  export let timeString: string | undefined = undefined;
  export let hideDivider = false;

  export let onClick: (range: string) => void;

  $: filtered = options.filter((option) => {
    return (
      option.interval.toString().toLowerCase().includes(filter.toLowerCase()) ||
      option.getLabel().toLowerCase().includes(filter.toLowerCase())
    );
  });
</script>

{#if filtered.length}
  <div class="w-full h-fit px-1">
    {#if hideDivider}
      <div class="h-px w-full bg-gray-200 my-1" />
    {/if}
    {#each filtered as option, i (i)}
      <TimeRangeMenuItem
        range={option.interval.toString()}
        selected={!!timeString?.startsWith(option.interval.toString())}
        label={option.getLabel()}
        {onClick}
      />
    {/each}

    {#if !hideDivider}
      <div class="h-px w-full bg-gray-200 my-1" />
    {/if}
  </div>
{/if}
