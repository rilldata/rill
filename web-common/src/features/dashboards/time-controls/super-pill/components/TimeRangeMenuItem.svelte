<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import SyntaxElement from "./SyntaxElement.svelte";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";

  export let rillTime: RillTime;
  export let timeString: string | undefined;
  export let onClick: (range: string) => void;

  $: label = rillTime.getLabel();
  $: range = rillTime.interval.toString();

  $: selected = !!timeString?.startsWith(range);
</script>

<div
  role="presentation"
  class="group h-7 pr-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
  on:click={() => {
    onClick(range);
  }}
>
  <button
    class:font-bold={selected}
    class="truncate w-full text-left flex-shrink pl-2 h-full flex items-center gap-x-1"
  >
    {#if range === "calendar"}
      <Calendar size="15px" />
    {/if}
    {label}
  </button>

  {#if range}
    <div class="sr-only group-hover:not-sr-only">
      <SyntaxElement {range} />
    </div>
  {/if}
</div>
