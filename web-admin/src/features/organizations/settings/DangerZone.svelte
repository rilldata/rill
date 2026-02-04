<script lang="ts">
  // Container component for dangerous/destructive actions
  // Styled similar to GitHub's "Danger Zone" section
</script>

<div class="danger-zone">
  <h2 class="danger-zone-header">Danger Zone</h2>
  <div class="danger-zone-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .danger-zone {
    @apply w-full rounded-sm overflow-hidden;
    border: 1px solid var(--color-red-400);
  }

  .danger-zone-header {
    @apply text-lg font-semibold;
    @apply px-5 py-3;
    color: var(--color-red-500);
  }

  .danger-zone-content {
    @apply flex flex-col;
  }

  .danger-zone-content > :global(*) {
    @apply border-t;
    border-color: var(--color-red-400);
  }

  /* Override the default border color for settings containers inside danger zone */
  .danger-zone-content > :global(.settings-container) {
    @apply border-0 rounded-none;
  }

  .danger-zone-content > :global(.settings-container .settings-footer) {
    @apply border-t;
    border-color: var(--color-red-400);
  }
</style>
