<!--
  Composer toggle: "Don't remember this chat".
  When on, the conversation is excluded from memory formation; existing memories are still used.
  The flag is sent with the next message and persisted on the conversation by the runtime.
-->
<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Brain, BrainCog } from "lucide-svelte";
  import type { Conversation } from "../core/conversation";
  import { useMemoryEnabled } from "./memory-store";

  export let conversation: Conversation;

  const client = useRuntimeClient();
  const enabled = useMemoryEnabled(client);

  $: memoryDisabled = conversation.memoryDisabled;
  $: conversationQuery = conversation.getConversationQuery();

  // Reflect the persisted flag when a conversation loads. Only sync once per conversation so a toggle
  // made before the next message is not overwritten by a refetch.
  let syncedConversationId: string | undefined;
  $: {
    const loaded = $conversationQuery.data?.conversation;
    if (loaded?.id && loaded.id !== syncedConversationId) {
      syncedConversationId = loaded.id;
      memoryDisabled.set(!!loaded.memoryDisabled);
    }
  }
</script>

{#if $enabled}
  <IconButton
    ariaLabel={m.chat_memory_dont_remember_chat()}
    ariaPressed={$memoryDisabled}
    active={$memoryDisabled}
    onclick={() => memoryDisabled.update((v) => !v)}
  >
    {#if $memoryDisabled}
      <BrainCog size="16px" class="text-fg-secondary" />
    {:else}
      <Brain size="16px" class="text-fg-muted" />
    {/if}
    <svelte:fragment slot="tooltip-content">
      {$memoryDisabled
        ? m.chat_memory_dont_remember_chat_on()
        : m.chat_memory_dont_remember_chat_off()}
    </svelte:fragment>
  </IconButton>
{/if}
