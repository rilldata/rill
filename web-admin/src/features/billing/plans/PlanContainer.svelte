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
    footer,
  }: {
    badge: string;
    description: string;
    info?: Snippet;

    action: Snippet;
    children: Snippet;
    footer?: Snippet;
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

    {#if footer}
      <div class="plan-footer">
        {@render footer()}
      </div>
    {/if}
  </div>
</section>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary;
  }

  .plan-card {
    @apply border rounded-xl bg-surface-background p-6 shadow-md;
  }

  .plan-top {
    @apply flex items-center justify-between;
  }

  .plan-badge {
    @apply inline-flex items-center justify-center px-2 h-[21px] rounded-full border-none;
    @apply text-xs font-semibold bg-surface-muted;
  }

  .plan-description {
    @apply font-sans font-semibold text-lg leading-7 align-middle text-fg-tertiary;
  }

  .plan-footer {
    @apply flex items-center justify-between bg-surface-subtle border-t rounded-b-xl;
    margin: 16px -24px -24px;
    padding: 12px 24px;
  }
</style>
