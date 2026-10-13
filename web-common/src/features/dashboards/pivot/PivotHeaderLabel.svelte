<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let label: string;
  export let description: string | undefined = undefined;
  // Wrap the label over `--wrap-lines` lines instead of truncating it.
  export let wrap = false;
</script>

<Tooltip location="top" distance={8} activeDelay={200}>
  <p class:truncate={!wrap} class:wrap-text={wrap}>{label}</p>
  <TooltipContent slot="tooltip-content" maxWidth="280px">
    <div class="pointer-events-none items-baseline" aria-label="tooltip-name">
      {label}
    </div>
    {#if description}
      <div
        class="text-fg-inverse/70 pointer-events-none pt-1"
        style:max-width="280px"
        aria-label="tooltip-name-description"
      >
        {description}
      </div>
    {/if}
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  .wrap-text {
    @apply whitespace-normal overflow-hidden;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: var(--wrap-lines, 2);
    overflow-wrap: anywhere;
    line-height: 1rem;
  }
</style>
