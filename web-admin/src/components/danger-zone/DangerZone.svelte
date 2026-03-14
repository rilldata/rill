<!-- Container component for dangerous/destructive actions.
     Styled with red border and footer similar to Vercel's danger zones. -->
<div class="danger-zone">
  <slot />
</div>

<style lang="postcss">
  .danger-zone {
    @apply w-full rounded-sm overflow-hidden;
    border: 1px solid var(--color-red-200);
  }

  /* Override the default styling for settings containers inside danger zone */
  .danger-zone > :global(.settings-container) {
    @apply border-0 rounded-none;
  }

  /* Style the footer with a red-tinted background */
  .danger-zone :global(.settings-container .settings-footer) {
    @apply border-t;
    border-color: var(--color-red-200);
    background-color: color-mix(
      in oklab,
      var(--color-red-100) 40%,
      transparent
    );
  }
</style>
