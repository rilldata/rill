<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
  import { developerChatConfig } from "@rilldata/web-common/features/editor/chat-utils.ts";

  const { developerChat } = featureFlags;
</script>

{#if $developerChat && $chatOpen}
  <SidebarChat config={developerChatConfig} />
{/if}
