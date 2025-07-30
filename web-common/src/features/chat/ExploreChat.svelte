<script lang="ts">
  import ChatSidebar from "./layouts/sidebar/ChatSidebar.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
</script>

{#if $chatOpen}
  <ChatSidebar />
{/if}
