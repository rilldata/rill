<script lang="ts">
  export let text: string =
    "AI can make mistakes. Consider your dashboard the source of truth.";
</script>

<div class="chat-footer bg-surface-subtle">
  <div class="chat-footer-text">
    {text}
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
