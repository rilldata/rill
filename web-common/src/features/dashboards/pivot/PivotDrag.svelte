<script context="module" lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragList from "./DragList.svelte";
  import type {
    PivotChipData,
    PivotSidebarSection,
    PivotTableMode,
  } from "./types";
</script>

<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  // `title` is the stable, non-localized section identifier. It doubles as the
  // drag `zone` key (DragList compares zone === "Time"/"Measures"/"Dimensions"),
  // so it must NOT be localized. `label` is the localized header shown to the
  // user and defaults to `title` when omitted.
  export let title: PivotSidebarSection;
  export let label: string = title;
  export let items: PivotChipData[];
  export let collapsed = false;
  export let tableMode: PivotTableMode = "nest";

  function toggleCollapse() {
    collapsed = !collapsed;
  }
</script>

<div class="container">
  <button
    class="flex gap-1 w-full items-start flex-none"
    onclick={toggleCollapse}
  >
    <span class="header">{label}</span>
    <div class="transition-transform" class:-rotate-180={!collapsed}>
      <CaretDownIcon size="12px" />
    </div>
  </button>

  {#if !collapsed}
    <div class="w-full h-fit overflow-x-hidden px-[2px] mt-2">
      {#if items.length}
        <DragList {items} zone={title} {tableMode} />
      {:else}
        <p class="text-fg-secondary my-1">
          {m.dashboard_no_available_fields()}
        </p>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .container {
    @apply py-3 px-4;
    @apply flex flex-col gap-1 items-start;
    @apply w-full;
    @apply border-b h-fit;
  }

  .container:last-of-type {
    @apply border-b-0;
  }

  .header {
    @apply uppercase font-semibold text-[10px];
  }
</style>
