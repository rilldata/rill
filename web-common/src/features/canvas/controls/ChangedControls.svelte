<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // Ends a level's control bar when some of its controls are off their defaults, and lets
  // viewers go back. It sits on the bar's line, so showing it doesn't move the widgets below.
  let { onReset }: { onReset: () => void } = $props();
</script>

<div class="changed-controls">
  <span class="truncate">{m.canvas_controls_changed()}</span>
  <button type="button" onclick={onReset}>
    {m.canvas_controls_reset_all()}
  </button>
</div>

<style lang="postcss">
  .changed-controls {
    @apply ml-auto flex min-w-0 shrink items-center gap-x-2;
    @apply text-xs text-fg-secondary;
  }

  button {
    @apply flex-none whitespace-nowrap text-primary-600 font-medium;
  }

  button:hover {
    @apply underline;
  }
</style>
