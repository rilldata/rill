<script lang="ts">
  import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
  import LocalFiltersHeader from "@rilldata/web-common/features/canvas/LocalFiltersHeader.svelte";

  export let title: string | undefined = undefined;
  export let description: string | undefined = undefined;
  export let filters: ComponentFilterProperties;
  export let faint: boolean = false;

  $: atleastOneFilter = Boolean(
    filters?.comparison_range ||
      filters?.dimension_filters ||
      filters?.time_range,
  );
</script>

{#if title || description}
  <div class="w-full h-fit flex flex-col bg-white p-2 items-start">
    {#if title}
      <div class="flex items-center gap-x-2">
        <h1 class:faint class="title">{title}</h1>
        <LocalFiltersHeader {atleastOneFilter} {filters} />
      </div>
    {/if}
    {#if description}
      <div class="flex items-center gap-x-2">
        <h2 class="description">{description}</h2>
        {#if !title}
          <LocalFiltersHeader {atleastOneFilter} {filters} />
        {/if}
      </div>
    {/if}
  </div>
{:else if atleastOneFilter}
  <div class="absolute top-0 left-0 z-50 px-2 py-1">
    <LocalFiltersHeader {atleastOneFilter} {filters} />
  </div>
{/if}

<style lang="postcss">
  h1 {
    font-size: 16px;
    font-weight: 500;
  }

  h2 {
    font-size: 12px;
    font-weight: 400;
  }
  .title {
    font-size: 15px;
    @apply font-medium text-slate-700 truncate;
  }

  .title.faint {
    @apply text-gray-500;
  }

  .description {
    font-size: 13px;
    @apply text-slate-600 leading-none;
  }
</style>
