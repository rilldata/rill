<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import {
    developerChatActions,
    developerChatOpen,
  } from "./layouts/sidebar/sidebar-store";
  import { developerChatConfig } from "@rilldata/web-common/features/editor/chat-utils.ts";

  const { developerChat } = featureFlags;
</script>

{#if $developerChat && $developerChatOpen}
  <SidebarChat
    config={developerChatConfig}
    actions={developerChatActions}
    surface="developer"
  />
{/if}
