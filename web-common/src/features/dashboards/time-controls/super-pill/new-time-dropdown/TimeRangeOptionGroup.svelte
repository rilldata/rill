<script lang="ts">
  import TimeRangeMenuItem from "../components/TimeRangeMenuItem.svelte";
  import type { TimeRangeMenuOption } from "@rilldata/web-common/lib/time/defaults";

  export let filter: string;
  export let options: TimeRangeMenuOption[];
  export let timeString: string | undefined;
  export let type: "this" | "last" | "ago" | "by";
  export let onClick;

  $: filtered = options.filter(({ string, parsed }) => {
    return (
      parsed.getLabel().toLowerCase().includes(filter.toLowerCase()) ||
      string.toLowerCase().includes(filter.toLowerCase())
    );
  });
</script>

{#each filtered as { string, parsed }, i (i)}
  <TimeRangeMenuItem
    range={string}
    {type}
    selected={timeString === string}
    {parsed}
    {onClick}
  />
{/each}

{#if filtered.length}
  <div class="h-px w-full bg-gray-200 my-1" />
{/if}
