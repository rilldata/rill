<script lang="ts">
  import type { Writable } from "svelte/store";
  import type { TimeDimensionDetailsStore } from "./time-dimension-details-store";
  import { getContext } from "svelte";

  export const rowIdx: number = -1;
  export let colIdx: number;
  export let fixed = false;
  export let lastFixed = false;
  const { store, headers } = getContext<{
    headers: { title: string }[];
    store: Writable<TimeDimensionDetailsStore>;
  }>("tdt-store");

  let _class = "";
  $: {
    _class = "h-full bg-white border-b text-left px-2 flex items-center";
    if (fixed) _class += ` z-2`;
    if (lastFixed) _class += ` right-shadow`;
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
  {header.title}
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
