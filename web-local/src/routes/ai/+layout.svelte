<script lang="ts">
  import AiChatPage from "@rilldata/web-common/features/preview-mode/AiChatPage.svelte";
</script>

<div class="chat-page-wrapper">
  <AiChatPage basePath="/ai" />
  <!-- This slot isn't used, but its presence avoids a SvelteKit browser console warning. -->
  <slot />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    height: 100%;
    width: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: var(--surface-base);
  }
</style>
