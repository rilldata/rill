<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import DimensionCompareMenu from "../leaderboard/DimensionCompareMenu.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";

  const dispatch = createEventDispatcher();

  export let index: number;
  export let width: string;
  export let sorted: 1 | 2 | undefined;
  export let dimensionName: string;
  export let label: string;
  export let sortedAscending: boolean;
  export let name: string;
  export let description: string | undefined;
</script>

<!-- <th
  class:first-column-header={index === 0}
  style:width
  style:max-width={width}
  class:sorted
  class="py-1"
  data-description={description}
  data-index={index}
  data-header={true}
  data-column={name}
  on:mouseenter
> -->
<button
  class:text-left={index === 0}
  class="header-cell size-full px-2 text-right"
  class:sorted
>
  <!-- {#if index === 0}
    <span class="bg-white h-full flex items-center">
      <DimensionCompareMenu {dimensionName} />
    </span>
  {/if} -->
  <!-- <p
    class:text-right={index !== 0}
    class="w-full text-left break-words line-clamp-2 outline"
  > -->

  <p
    class="break-words whitespace-normal line-clamp-2 w-full font-normal"
    class:sorted
  >
    {label}
  </p>
  <!-- </p> -->
  {#if sorted}
    <div class="mt-0.5 ui-copy-icon" class:-rotate-180={sorted === 1}>
      <ArrowDown size="12px" />
    </div>
  {/if}
</button>
<slot name="resizer" />

<!-- </th> -->

<style lang="postcss">
  .header-cell {
    @apply flex flex-row gap-x-1 size-full items-center pr-1;
  }

  .sorted {
    @apply font-bold;
  }

  .first-column-header {
    @apply sticky left-0;
    @apply z-10 text-left;
    @apply truncate border-r;
  }

  .sorted {
    @apply bg-gray-50;
  }

  th {
    @apply text-right relative bg-white;
    font-weight: inherit;
  }
</style>
