<script lang="ts">
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";
</script>

<div class="chat-page-wrapper">
  <ProjectChat />
  <slot />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    @apply bg-surface-background;
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
</style>
