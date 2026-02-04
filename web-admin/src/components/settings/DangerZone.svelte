<script lang="ts">
  // Container component for dangerous/destructive actions
  // Styled with light red border and red-tinted footer similar to GitHub danger zones
  // Groups multiple items with internal dividers
</script>

<div class="danger-zone">
  <slot />
</div>

<style lang="postcss">
  .danger-zone {
    @apply w-full rounded-sm overflow-hidden;
    border: 1px solid var(--color-red-200);
  }

  /* Override the default border color for settings containers inside danger zone */
  .danger-zone > :global(.settings-container) {
    @apply border-0 rounded-none;
  }

  /* Add divider between multiple settings containers */
  .danger-zone > :global(.settings-container + .settings-container) {
    border-top: 1px solid var(--color-red-200);
  }

  .danger-zone > :global(.settings-container .settings-footer) {
    @apply border-t;
    border-color: var(--color-red-200);
    background-color: color-mix(in srgb, var(--color-red-400) 10%, transparent);
  }
</style>
