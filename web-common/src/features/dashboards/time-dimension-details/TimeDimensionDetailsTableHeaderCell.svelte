<script lang="ts">
  import { useTDTContext } from "./context";

  export const rowIdx = -1;
  export let colIdx: number;
  export let fixed = false;
  export let lastFixed = false;

  const { store, headers } = useTDTContext();

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
