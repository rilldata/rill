<!--
  Inline card asking the user to authorize an external MCP server the assistant tried to use.
  After the popup completes, a follow-up message is sent so the assistant picks up the same task.
  The card stays visible in the history; once the server is connected it only offers to continue.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlugZap } from "lucide-svelte";
  import { get } from "svelte/store";
  import { getMcpConnectContext } from "../../../connect/mcp-connect-context";
  import {
    createMcpServersQuery,
    invalidateMcpServers,
    openMcpConnectPopup,
  } from "../../../connect/mcp-servers";
  import type { Conversation } from "../../conversation";
  import type { ChatConfig } from "../../types";
  import type { McpAuthRequiredBlock } from "./mcp-auth-required-block";

  export let block: McpAuthRequiredBlock;
  export let conversation: Conversation;
  export let config: ChatConfig;

  const runtimeClient = useRuntimeClient();
  const connectContext = getMcpConnectContext();
  const serversQuery = createMcpServersQuery(runtimeClient);

  let connecting = false;
  let error: string | null = null;

  // Whether this card is the latest turn of the conversation, in which case a successful
  // connection resumes the conversation automatically.
  $: blocksStore = conversation.getBlocks();
  $: isLatest = $blocksStore[$blocksStore.length - 1]?.id === block.id;
  $: isStreaming = conversation.isStreaming;
  $: server = ($serversQuery.data ?? []).find(
    (s) => s.connector === block.connector,
  );
  $: alreadyConnected = !!server?.connected;

  async function connect() {
    if (!connectContext) return;
    connecting = true;
    error = null;
    const result = await openMcpConnectPopup(
      connectContext.connectUrl(block.connector),
      block.connector,
    );
    connecting = false;
    await invalidateMcpServers(runtimeClient);
    if (result.status === "connected") {
      await resume();
    } else if (result.status === "error") {
      error = result.error ?? m.chat_mcp_status_error();
    }
  }

  // Sends a follow-up user message so the assistant retries with the server connected.
  // This keeps the runtime stateless about the interrupted call. Anything the user had typed is kept.
  async function resume() {
    if (get(isStreaming)) return;
    const draft = get(conversation.draftMessage);
    conversation.draftMessage.set(
      m.chat_mcp_continue_message({ name: block.displayName }),
    );
    const additionalContext = get(config.additionalContextStoreGetter());
    try {
      await conversation.sendMessage(additionalContext, {
        onStreamStart: () => conversation.draftMessage.set(draft),
      });
    } catch (err) {
      conversation.draftMessage.set(draft);
      console.error("Failed to resume the conversation:", err);
    }
  }
</script>

<div class="mcp-auth-card" role="status">
  <div class="mcp-auth-icon">
    <PlugZap size="16px" />
  </div>
  <div class="mcp-auth-body">
    <div class="mcp-auth-title">
      {m.chat_mcp_auth_required_title({ name: block.displayName })}
    </div>
    <p class="mcp-auth-text">
      {#if alreadyConnected}
        {m.chat_mcp_auth_connected_body({ name: block.displayName })}
      {:else if connectContext}
        {m.chat_mcp_auth_required_body({ name: block.displayName })}
      {:else}
        {m.chat_mcp_auth_required_no_provider({ name: block.displayName })}
      {/if}
    </p>
    {#if error}
      <p class="mcp-auth-error">{error}</p>
    {/if}
    {#if isLatest}
      <div class="mcp-auth-actions">
        {#if alreadyConnected}
          <Button type="primary" small onClick={resume} disabled={$isStreaming}>
            {m.chat_mcp_continue()}
          </Button>
        {:else if connectContext}
          <Button
            type="primary"
            small
            onClick={connect}
            loading={connecting}
            disabled={connecting}
          >
            {m.chat_mcp_connect_server({ name: block.displayName })}
          </Button>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .mcp-auth-card {
    @apply flex w-full max-w-full items-start gap-x-3 self-start;
    @apply rounded-md border border-primary-200 bg-primary-50 p-3;
  }

  .mcp-auth-icon {
    @apply mt-0.5 shrink-0 text-primary-600;
  }

  .mcp-auth-body {
    @apply flex min-w-0 flex-col gap-y-1;
  }

  .mcp-auth-title {
    @apply text-sm font-medium text-fg-primary;
  }

  .mcp-auth-text {
    @apply text-xs text-fg-secondary;
  }

  .mcp-auth-error {
    @apply text-xs text-red-600;
  }

  .mcp-auth-actions {
    @apply mt-1 flex gap-x-2;
  }
</style>
