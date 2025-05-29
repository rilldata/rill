<script lang="ts">
  import TimeRangeMenuItem from "../components/TimeRangeMenuItem.svelte";
  import type { TimeRangeMenuOption } from "@rilldata/web-common/lib/time/defaults";

  export let filter = "";
  export let options: TimeRangeMenuOption[];
  export let timeString: string | undefined;
  export let onClick: (range: string, syntax: boolean) => void;

  $: filtered = options.filter(({ string, label }) => {
    return (
      label.toLowerCase().includes(filter.toLowerCase()) ||
      string.toLowerCase().includes(filter.toLowerCase())
    );
  });
</script>

<div class="w-full h-fit px-1">
  {#each filtered as { string, label }, i (i)}
    <TimeRangeMenuItem
      range={string}
      selected={timeString === string}
      {label}
      {onClick}
    />
  {/each}
</div>

{#if filtered.length}
  <div class="h-px w-full bg-gray-200 my-1" />
{/if}
