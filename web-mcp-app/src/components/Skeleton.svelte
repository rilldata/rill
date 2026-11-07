<script lang="ts">
  let {
    height = 300,
    compact = false,
    label,
  }: { height?: number; compact?: boolean; label: string } = $props();
</script>

<div
  class="skeleton"
  class:compact
  role="status"
  aria-live="polite"
  aria-label={label}
  style:height={compact ? undefined : `${height + 96}px`}
>
  {#if !compact}
    <div class="line title"></div>
    <div class="line caption"></div>
  {/if}
  <div class="bars" class:compact>
    <div class="bar" style:height="55%"></div>
    <div class="bar" style:height="80%"></div>
    <div class="bar" style:height="40%"></div>
    <div class="bar" style:height="65%"></div>
    <div class="bar" style:height="30%"></div>
    <div class="bar" style:height="50%"></div>
  </div>
</div>

<style lang="postcss">
  .skeleton {
    @apply flex w-full flex-col gap-2 px-4 py-3;
    animation: pulse 1.4s ease-in-out infinite;
  }
  .skeleton.compact {
    @apply w-1/2 max-w-xs py-0;
  }
  .line {
    @apply h-3 rounded-full bg-surface-active;
  }
  .line.title {
    @apply w-2/5 h-4;
  }
  .line.caption {
    @apply w-3/5;
  }
  .bars {
    @apply mt-2 flex flex-1 items-end gap-2 min-h-[120px];
  }
  .bars.compact {
    @apply min-h-[80px];
  }
  .bar {
    @apply flex-1 rounded-t bg-surface-active;
  }
  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.45;
    }
  }
</style>
