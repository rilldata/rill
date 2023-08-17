<script lang="ts">
  import { FILTER_OVERFLOW_WIDTH } from "./constants";
  import { useTDDContext } from "./context";

  export const rowIdx = -1;
  export let colIdx: number;
  export let fixed = false;
  export let lastFixed = false;

  const { store, headers } = useTDDContext();

  let _class = "";
  $: {
    _class =
      "h-full w-full bg-white border-b px-2 flex items-center overflow-hidden";
    if (fixed) _class += ` z-2`;
    if (lastFixed) _class += ` right-shadow`;
    if (colIdx > 0) _class += ` justify-end font-medium`;
    else _class += ` justify-start font-bold`;
  }

  const handleMouseEnter = () => {
    $store.highlightedCol = colIdx;
  };
  const handleMouseLeave = () => {
    $store.highlightedCol = null;
  };

  $: header = headers[colIdx];
</script>

<div
  class={_class}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
>
  {#if header.component}
    <svelte:component this={header.component} />
  {:else}
    {header.title}
  {/if}
  {#if colIdx === 0}
    <!-- Placeholder for table padding to fit the overflowing checkboxes from dimension cells. This will hide column labels that horizontally scroll to the left -->
    <div
      style={`left: -${FILTER_OVERFLOW_WIDTH}px; width: ${FILTER_OVERFLOW_WIDTH}px;`}
      class="absolute top-0 h-full bg-white z-10 flex items-center justify-center bg-white border-b"
    />
  {/if}
</div>

<style>
  .right-shadow:after {
    content: "";
    width: 1px;
    height: 100%;
    position: absolute;
    top: 0px;
    right: 0px;
    background: #e5e7eb;
    filter: drop-shadow(3px 0px 3px rgb(0 0 0 / 0.27));
  }
</style>
