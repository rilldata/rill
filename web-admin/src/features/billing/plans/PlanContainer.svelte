<script lang="ts">
  import type { Snippet } from "svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";

  let {
    badge,
    description,
    info,

    action,
    children,
  }: {
    badge: string;
    description: string;
    info?: Snippet<[]>;

    action: Snippet;
    children: Snippet;
  } = $props();
</script>

<section>
  <h2 class="section-header">Plan</h2>

  <div class="plan-card">
    <div class="plan-top">
      <div class="flex items-center gap-3">
        <div class="plan-badge">{badge}</div>
        <div class="plan-description">{description}</div>
        {#if info}
          <Tooltip location="right" alignment="middle" distance={8}>
            <span class="text-fg-muted flex">
              <InfoCircle size="16px" />
            </span>
            <TooltipContent maxWidth="240px" slot="tooltip-content">
              {@render info()}
            </TooltipContent>
          </Tooltip>
        {/if}
      </div>
      <div class="flex items-center gap-2">
        {@render action()}
      </div>
    </div>

    {@render children?.()}
  </div>
</section>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }

  .plan-card {
    @apply border rounded-xl bg-surface-background p-6 shadow-md;
  }

  .plan-top {
    @apply flex items-center justify-between;
  }

  .plan-badge {
    @apply inline-flex items-center justify-center w-[76px] h-[21px] rounded-full border-none;
    @apply text-xs font-semibold bg-surface-muted;
  }

  .plan-description {
    @apply font-sans font-semibold text-lg leading-7 align-middle text-fg-tertiary;
  }
</style>
