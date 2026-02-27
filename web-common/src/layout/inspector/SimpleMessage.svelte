<script lang="ts">
  import DOMPurify from "dompurify";

  export let message: string;
  export let includesHtml = false;
</script>

<div class="custom-instructions-wrapper" style:text-wrap="balance">
  {#if includesHtml}
    <p>{@html DOMPurify.sanitize(message)}</p>
  {:else}
    <p>{message}</p>
  {/if}
</div>

<style lang="postcss">
  .custom-instructions-wrapper {
    @apply px-4 py-24 w-full;
    @apply italic text-fg-secondary text-center;
  }
</style>
