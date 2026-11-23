<script lang="ts">
  export interface SegmentedOption {
    value: string;
    label: string;
  }

  let {
    options,
    value,
    label,
    onchange,
  }: {
    options: SegmentedOption[];
    value: string | undefined;
    /** Accessible group name; also rendered as a small leading label. */
    label: string;
    onchange: (value: string) => void;
  } = $props();
</script>

<div class="segmented-group">
  <span class="group-label" aria-hidden="true">{label}</span>
  <div class="segmented" role="radiogroup" aria-label={label}>
    {#each options as option (option.value)}
      <button
        type="button"
        role="radio"
        aria-checked={option.value === value}
        class="segment"
        class:selected={option.value === value}
        onclick={() => onchange(option.value)}
      >
        {option.label}
      </button>
    {/each}
  </div>
</div>

<style lang="postcss">
  .segmented-group {
    @apply inline-flex items-center gap-2;
  }
  .group-label {
    @apply text-xs text-fg-tertiary whitespace-nowrap;
  }
  .segmented {
    @apply inline-flex rounded-[var(--radius)] border border-border bg-surface-subtle p-0.5;
  }
  .segment {
    /* 44px touch targets, visually compact through tight horizontal padding. */
    @apply min-h-[40px] px-3 rounded-[calc(var(--radius)-2px)] text-[13px] font-medium;
    @apply bg-transparent text-fg-secondary cursor-pointer transition-colors whitespace-nowrap;
    @apply hover:text-fg-primary;
  }
  .segment.selected {
    @apply bg-surface-background text-fg-primary shadow-sm;
  }
</style>
