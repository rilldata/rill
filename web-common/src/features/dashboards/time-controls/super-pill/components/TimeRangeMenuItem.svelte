<script lang="ts">
  import SyntaxElement from "./SyntaxElement.svelte";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";

  export let rillTime: RillTime;
  export let timeString: string | undefined;
  export let onClick: (range: string) => void;

  $: label = rillTime.getLabel();
  $: range = rillTime.interval.toString();

  $: selected = !!timeString?.startsWith(range);
</script>

<button
  type="button"
  role="menuitem"
  class:font-bold={selected}
  class="group truncate h-7 p-2 justify-between overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
  on:click={() => {
    onClick(range);
  }}
>
  {label}

  {#if range}
    <div class="sr-only group-hover:not-sr-only">
      <SyntaxElement {range} />
    </div>
  {/if}
</button>
