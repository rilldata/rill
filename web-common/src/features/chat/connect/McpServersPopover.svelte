<!--
  Lists the external MCP servers the current user may use in this project, with their authorization status.
  This is the one place where users connect (or disconnect) OAuth servers; the Connect card in a conversation opens it.
  Renders nothing when the project has no external servers.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Plug } from "lucide-svelte";
  import { getMcpConnectContext } from "./mcp-connect-context";
  import {
    createDisconnectMcpServerMutation,
    createMcpServersQuery,
    invalidateMcpServers,
    mcpServersPopoverOpen,
    openMcpConnectPopup,
    type McpServer,
  } from "./mcp-servers";

  const runtimeClient = useRuntimeClient();
  const connectContext = getMcpConnectContext();
  const serversQuery = createMcpServersQuery(runtimeClient);
  const disconnectMutation = createDisconnectMcpServerMutation(runtimeClient);

  let connecting: string | null = null;
  let connectError: string | null = null;

  $: servers = $serversQuery.data ?? [];
  $: unconnectedCount = servers.filter((s) => s.oauth && !s.connected).length;

  async function connect(server: McpServer) {
    if (!connectContext || !server.connector) return;
    connecting = server.connector;
    connectError = null;
    const result = await openMcpConnectPopup(
      connectContext.connectUrl(server.connector),
      server.connector,
    );
    connecting = null;
    if (result.status === "error") {
      connectError = m.chat_mcp_connect_failed({
        name: server.displayName ?? server.connector,
        error: result.error ?? "",
      });
    }
    await invalidateMcpServers(runtimeClient);
  }

  function disconnect(server: McpServer) {
    if (!server.connector) return;
    $disconnectMutation.mutate({ connector: server.connector });
  }

  // A server is usable only when its tools could be listed for the user. Without OAuth there is nothing
  // the user can do about it, so it shows as unavailable rather than as connected.
  function statusLabel(server: McpServer): string {
    if (server.error) return m.chat_mcp_status_error();
    if (!server.connected) {
      return server.oauth
        ? m.chat_mcp_status_not_connected()
        : m.chat_mcp_status_error();
    }
    return m.chat_mcp_status_connected();
  }

  function isUsable(server: McpServer): boolean {
    return !server.error && !!server.connected;
  }
</script>

{#if servers.length > 0}
  <Popover bind:open={$mcpServersPopoverOpen}>
    <PopoverTrigger>
      <IconButton
        ariaLabel={m.chat_mcp_servers()}
        bgGray
        active={$mcpServersPopoverOpen}
        disableTooltip={$mcpServersPopoverOpen}
      >
        <span class="relative flex">
          <Plug size="16px" class="text-fg-muted" />
          {#if unconnectedCount > 0}
            <span class="mcp-badge" aria-hidden="true"></span>
          {/if}
        </span>
        <svelte:fragment slot="tooltip-content">
          {m.chat_mcp_servers()}
        </svelte:fragment>
      </IconButton>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-[320px] p-4">
      <div class="flex flex-col gap-y-3">
        <div class="flex flex-col gap-y-1">
          <h3 class="text-sm font-medium text-fg-primary">
            {m.chat_mcp_servers()}
          </h3>
          <p class="text-xs text-fg-secondary">
            {m.chat_mcp_servers_description()}
          </p>
        </div>
        <ul class="mcp-server-list">
          {#each servers as server (server.connector)}
            <li class="mcp-server">
              <div class="mcp-server-info">
                <span class="mcp-server-name">
                  {server.displayName || server.connector}
                </span>
                <span
                  class="mcp-server-status"
                  class:connected={isUsable(server)}
                  class:error={!!server.error ||
                    (!server.connected && !server.oauth)}
                >
                  {statusLabel(server)}
                  {#if isUsable(server) && server.tools?.length}
                    · {m.chat_mcp_tools_count({
                      count: String(server.tools.length),
                    })}
                  {/if}
                </span>
                {#if server.error}
                  <span class="mcp-server-error" title={server.error}>
                    {server.error}
                  </span>
                {/if}
              </div>
              {#if server.oauth && connectContext}
                {#if server.connected}
                  <Button
                    type="text"
                    small
                    onClick={() => disconnect(server)}
                    disabled={$disconnectMutation.isPending}
                  >
                    {m.chat_mcp_disconnect()}
                  </Button>
                {:else}
                  <Button
                    type="secondary"
                    small
                    onClick={() => connect(server)}
                    loading={connecting === server.connector}
                    disabled={connecting !== null}
                  >
                    {m.chat_mcp_connect()}
                  </Button>
                {/if}
              {/if}
            </li>
          {/each}
        </ul>
        {#if connectError}
          <p class="text-xs text-red-600">{connectError}</p>
        {/if}
      </div>
    </PopoverContent>
  </Popover>
{/if}

<style lang="postcss">
  .mcp-badge {
    @apply absolute -top-0.5 -right-0.5 size-2 rounded-full bg-primary-500;
  }

  .mcp-server-list {
    @apply flex flex-col gap-y-2;
  }

  .mcp-server {
    @apply flex items-center justify-between gap-x-2;
  }

  .mcp-server-info {
    @apply flex min-w-0 flex-col;
  }

  .mcp-server-name {
    @apply truncate text-sm text-fg-primary;
  }

  .mcp-server-status {
    @apply text-xs text-fg-muted;
  }

  .mcp-server-status.connected {
    @apply text-fg-secondary;
  }

  .mcp-server-status.error {
    @apply text-red-600;
  }

  .mcp-server-error {
    @apply truncate text-xs text-fg-muted;
  }
</style>
