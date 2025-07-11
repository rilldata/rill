<script lang="ts">
  import SyntaxElement from "./SyntaxElement.svelte";

  export let range: string;
  export let label: string;
  export let selected: boolean;
  export let onClick: (range: string) => void;
</script>

<div
  role="presentation"
  on:click={() => {
    onClick(range);
  }}
  class="group h-7 pr-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
>
  <button
    class:font-bold={selected}
    class="truncate w-full text-left flex-shrink pl-2 h-full"
  >
    {label}
  </button>

  <div class="sr-only group-hover:not-sr-only">
    <SyntaxElement {range} />
  </div>
</div>
