<script lang="ts">
  import SyntaxElement from "./SyntaxElement.svelte";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";

  export let rillTime: RillTime;
  export let timeString: string | undefined;
  export let onClick: (range: string) => void;

  $: label = rillTime.getLabel();
  // If there is as of baked into the range then use it.
  $: range = rillTime.asOfLabel
    ? rillTime.toString()
    : rillTime.interval.toString();

  $: selected = !!timeString?.startsWith(range);
</script>

<button
  type="button"
  role="menuitem"
  class:font-bold={selected}
  class="group truncate h-7 p-2 text-popover-foreground justify-between overflow-hidden hover:bg-popover-accent rounded-sm w-full select-none flex items-center"
  onclick={() => {
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
