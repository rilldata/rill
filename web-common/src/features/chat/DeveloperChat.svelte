<script lang="ts">
  import { onDestroy } from "svelte";
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
  import { developerChatConfig } from "@rilldata/web-common/features/editor/chat-utils.ts";
  import { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
  import { maybeNavigateToGeneratedFile } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";

  const { developerChat } = featureFlags;

  let conversationManager: ConversationManager | null = null;
  let navigationUnsub: (() => void) | null = null;
  $: if (conversationManager) {
    navigationUnsub?.();
    navigationUnsub = maybeNavigateToGeneratedFile(conversationManager);
  }

  onDestroy(() => navigationUnsub?.());
</script>

{#if $developerChat && $chatOpen}
  <SidebarChat config={developerChatConfig} bind:conversationManager />
{/if}
