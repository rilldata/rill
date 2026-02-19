<script lang="ts">
  import LocalFullPageChat from "./LocalFullPageChat.svelte";
</script>

<div class="chat-page-wrapper">
  <LocalFullPageChat />
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
