<script lang="ts">
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
  import LocalFiltersHeader from "@rilldata/web-common/features/canvas/LocalFiltersHeader.svelte";
  import { onDestroy, onMount } from "svelte";

  export let title: string | undefined = undefined;
  export let description: string | undefined = undefined;
  export let showDescriptionAsTooltip: boolean | undefined = false;
  export let filters: ComponentFilterProperties | undefined = undefined;
  export let faint: boolean = false;
  export let component: BaseCanvasComponent<any>;

  const WIDTH_THRESHOLD = 480;

  let container: HTMLDivElement;
  let wide = false;
  let resizeObserver: ResizeObserver;

  $: atleastOneFilter = Boolean(
    filters?.time_filters || filters?.dimension_filters,
  );

  onMount(() => {
    resizeObserver = new ResizeObserver(([entry]) => {
      wide = entry.contentRect.width >= WIDTH_THRESHOLD;
    });
    if (container) resizeObserver.observe(container);
  });

  onDestroy(() => {
    if (resizeObserver && container) resizeObserver.unobserve(container);
  });
</script>

{#if title || description}
  <div
    bind:this={container}
    class="component-header-container w-full h-fit flex flex-col bg-card px-4 pt-2 pb-1 items-start {wide
      ? 'wide'
      : ''}"
  >
    {#if title}
      <div class="header-row">
        {#if showDescriptionAsTooltip && description}
          <Tooltip location="bottom" alignment="start">
            <div class="title-with-icon">
              <h1 class:faint class="title">{title}</h1>
              <InfoCircle className="text-gray-500" size="16px" />
            </div>
            <TooltipContent slot="tooltip-content">
              {description}
            </TooltipContent>
          </Tooltip>
        {:else}
          <h1 class:faint class="title">{title}</h1>
        {/if}
        {#if atleastOneFilter}
          <LocalFiltersHeader {component} />
        {/if}
      </div>
    {/if}
    {#if description && !showDescriptionAsTooltip}
      <div class="header-row">
        <h2 class="description">{description}</h2>
        {#if !title && atleastOneFilter}
          <LocalFiltersHeader {component} />
        {/if}
      </div>
    {:else if !title && showDescriptionAsTooltip && description}
      <div class="header-row">
        <Tooltip location="bottom" alignment="start" distance={4}>
          <InfoCircle className="text-gray-500" size="16px" />
          <TooltipContent slot="tooltip-content">
            {description}
          </TooltipContent>
        </Tooltip>
        {#if atleastOneFilter}
          <LocalFiltersHeader {component} />
        {/if}
      </div>
    {/if}
  </div>
{:else if atleastOneFilter}
  <div class="w-full px-2 py-1">
    <LocalFiltersHeader {component} />
  </div>
{/if}

<style lang="postcss">
  .header-row {
    @apply flex flex-col items-start gap-y-1 gap-x-2 w-full;
  }

  .component-header-container.wide .header-row {
    @apply flex-row items-center;
  }

  .title-with-icon {
    @apply flex items-center gap-x-1 flex-shrink-0;
  }

  .title {
    font-size: 15px;
    line-height: 26px;
    @apply flex-shrink-0;
    @apply font-medium text-gray-800 truncate;
  }

  .title.faint {
    @apply text-gray-500;
  }

  .description {
    font-size: 13px;
    @apply flex-shrink-0;
    @apply text-gray-500 font-normal leading-none;
  }
</style>
