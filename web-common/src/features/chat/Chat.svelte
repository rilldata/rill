<script lang="ts">
  import ChatFullPage from "./layouts/fullpage/ChatFullPage.svelte";
  import ChatSidebar from "./layouts/sidebar/ChatSidebar.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  export let layout: "sidebar" | "fullpage" = "sidebar";
</script>

{#if layout === "fullpage"}
  <ChatFullPage />
{:else if $chatOpen}
  <ChatSidebar />
{/if}
