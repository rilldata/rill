<script lang="ts">
  // Individual item within a DangerZone section
  // Displays a title, description, and action slot
  export let title: string;
  export let description: string;
</script>

<div class="danger-zone-item">
  <div class="item-content">
    <h3 class="item-title">{title}</h3>
    <p class="item-description">{description}</p>
  </div>
  <div class="item-action">
    <slot name="action" />
  </div>
</div>

<style lang="postcss">
  .danger-zone-item {
    @apply flex flex-row items-center justify-between gap-4;
    @apply px-5 py-4;
    @apply bg-surface-background;
  }

  .item-content {
    @apply flex flex-col gap-1;
  }

  .item-title {
    @apply text-sm font-semibold text-fg-primary;
  }

  .item-description {
    @apply text-sm text-fg-tertiary;
  }

  .item-action {
    @apply flex-shrink-0;
  }
</style>
