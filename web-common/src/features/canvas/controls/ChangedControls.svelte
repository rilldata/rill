<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // Ends a level's control bar when some of its controls are off their defaults, and lets
  // viewers go back. It sits on the bar's line, so showing it doesn't move the widgets below.
  // When the line is short, its text gives way first, so the controls don't fold into a menu.
  let { onReset }: { onReset: () => void } = $props();
</script>

<div class="changed-controls">
  <span class="truncate" title={m.canvas_controls_changed()}>
    {m.canvas_controls_changed()}
  </span>
  <button type="button" onclick={onReset}>
    {m.canvas_controls_reset_all()}
  </button>
</div>

<style lang="postcss">
  .changed-controls {
    @apply ml-auto flex min-w-0 items-center gap-x-2;
    @apply text-xs text-fg-secondary;
    flex: 0 1000 auto;
  }

  button {
    @apply flex-none whitespace-nowrap text-primary-600 font-medium;
  }

  button:hover {
    @apply underline;
  }
</style>
