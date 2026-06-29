<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let text: string | undefined = undefined;

  $: displayText = text ?? m.chat_ai_disclaimer();
</script>

<div class="chat-footer bg-surface-subtle">
  <div class="chat-footer-text">
    {displayText}
  </div>
</div>

<style lang="postcss">
  .chat-footer {
    padding: 0.5rem 1rem 1rem 1rem;
    font-size: 0.75rem;
  }

  .chat-footer-text {
    @apply text-fg-muted;
    text-align: center;
  }
</style>
