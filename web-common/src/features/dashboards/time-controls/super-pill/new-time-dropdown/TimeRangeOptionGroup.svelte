<script lang="ts">
  import TimeRangeMenuItem from "../components/TimeRangeMenuItem.svelte";
  import type { TimeRangeMenuOption } from "@rilldata/web-common/lib/time/defaults";

  export let filter = "";
  export let options: TimeRangeMenuOption[];
  export let timeString: string | undefined = undefined;
  export let hideDivider = false;
  export let onClick: (range: string) => void;

  $: filtered = options.filter(({ string, label }) => {
    return (
      label.toLowerCase().includes(filter.toLowerCase()) ||
      string.toLowerCase().includes(filter.toLowerCase())
    );
  });
</script>

<div class="w-full h-fit px-1">
  {#if hideDivider}
    <div class="h-px w-full bg-gray-200 my-1" />
  {/if}
  {#each filtered as { string, label }, i (i)}
    <TimeRangeMenuItem
      range={string}
      selected={!!timeString?.startsWith(string)}
      {label}
      {onClick}
    />
  {/each}

  {#if filtered.length && !hideDivider}
    <div class="h-px w-full bg-gray-200 my-1" />
  {/if}
</div>
