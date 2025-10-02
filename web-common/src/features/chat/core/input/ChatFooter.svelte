<script lang="ts">
  export let text: string =
    "AI can make mistakes. Consider your dashboard the source of truth.";
</script>

<div class="chat-footer">
  <div class="chat-footer-text">
    {text}
  </div>
</div>

<style lang="postcss">
  .chat-footer {
    padding: 0.5rem 1rem 1rem 1rem;
    background: var(--surface);
    font-size: 0.75rem;
  }

  .chat-footer-text {
    color: #6b7280;
    text-align: center;
  }
</style>
