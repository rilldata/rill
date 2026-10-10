<!--
  Chips shown above the chat input for OAuth servers the user has not connected yet,
  so they can connect before asking a question that needs the server's tools.
  Renders nothing when everything is connected or no provider can start the flow.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlugZap } from "lucide-svelte";
  import { getMcpConnectContext } from "./mcp-connect-context";
  import {
    createMcpServersQuery,
    invalidateMcpServers,
    openMcpConnectPopup,
    type McpServer,
  } from "./mcp-servers";

  /** Maximum number of chips before the rest is summarized. */
  const MAX_CHIPS = 3;

  const runtimeClient = useRuntimeClient();
  const connectContext = getMcpConnectContext();
  const serversQuery = createMcpServersQuery(runtimeClient);

  let connecting: string | null = null;

  $: unconnected = ($serversQuery.data ?? []).filter(
    (s) => s.oauth && !s.connected && !s.error,
  );
  $: visible = unconnected.slice(0, MAX_CHIPS);
  $: hiddenCount = unconnected.length - visible.length;

  async function connect(server: McpServer) {
    if (!connectContext || !server.connector) return;
    connecting = server.connector;
    await openMcpConnectPopup(
      connectContext.connectUrl(server.connector),
      server.connector,
    );
    connecting = null;
    await invalidateMcpServers(runtimeClient);
  }
</script>

{#if connectContext && visible.length > 0}
  <div class="mcp-chips">
    {#each visible as server (server.connector)}
      <Button
        type="secondary"
        small
        onClick={() => connect(server)}
        loading={connecting === server.connector}
        disabled={connecting !== null}
      >
        <PlugZap size="14px" />
        {m.chat_mcp_connect_server({
          name: server.displayName || server.connector || "",
        })}
      </Button>
    {/each}
    {#if hiddenCount > 0}
      <span class="text-xs text-fg-muted">
        {m.chat_mcp_more_servers({ count: String(hiddenCount) })}
      </span>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .mcp-chips {
    @apply flex flex-wrap items-center gap-1.5 px-4 pb-2;
  }
</style>
