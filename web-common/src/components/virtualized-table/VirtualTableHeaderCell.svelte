<script lang="ts">
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import type VirtualTableHeaderCell from "./VirtualTableHeaderCellContent.svelte";
  import type { ComponentType } from "svelte";
  import { HEADER_HEIGHT } from "./VirtualTable.svelte";

  export let HeaderCell: ComponentType<VirtualTableHeaderCell>;
  export let index: number;
  export let sorted: boolean;
  export let type: string | undefined;
  export let name: string | undefined;
  export let lastPinned: boolean = false;
  export let position: number | null = null;
  export let pinned: boolean = false;
  export let resizable: boolean = false;
</script>

<th
  id="header-{index}"
  data-index={index}
  data-column={name}
  class:last-pinned={lastPinned}
  class:pinned
  class="group relative overflow-hidden"
  style:left="{position}px"
  on:mouseenter
>
  <svelte:component this={HeaderCell} {sorted} {name} {type}>
    <button
      data-index={index}
      data-column={name}
      slot="pin-button"
      class="transition-colors duration-100 justify-self-end group-hover:block hidden text-gray-900"
      on:mouseenter
      on:click
    >
      <Pin size="16px" />
    </button>
  </svelte:component>

  {#if resizable && !pinned}
    <button
      class="absolute top-0 -right-1 w-2 z-10 cursor-col-resize"
      style:height="{HEADER_HEIGHT}px"
      data-index={index}
      on:mousedown
    />
  {/if}
</th>

<style lang="postcss">
  th {
    @apply truncate p-0 bg-white;
  }

  th:nth-last-of-type(2) {
    @apply border-r-0;
  }

  th.pinned {
    @apply z-50;
  }

  :global(.sticky-borders th) {
    @apply border-b;
  }

  :global(.header-borders th) {
    @apply border-r;
  }

  .pinned {
    @apply sticky;
  }
</style>
