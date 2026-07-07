<!-- 
  This layout wraps the chat page to provide proper height constraints
-->
<script lang="ts">
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";

  // The connect-to-AI-client entry points (empty-state hero + sidebar footer
  // button) and the MCPConnectDialog are provided by ConnectClientProvider in
  // the project layout, so this page only renders the chat.
</script>

<div class="chat-page-wrapper">
  <ProjectChat />

  <!-- This slot isn't used, but its presence avoids a SvelteKit browser console warning. -->
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
