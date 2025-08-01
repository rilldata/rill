<!-- 
  This layout wraps the chat page to provide proper height constraints
-->
<script lang="ts">
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";
</script>

<div class="chat-page-wrapper">
  <ProjectChat />

  <!-- This slot isn't used, but its presence avoids a SvelteKit browser console warning. -->
  <slot />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: #ffffff;
    min-height: 0;
  }
</style>
