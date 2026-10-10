<!--
  Inline card shown when the assistant tried to use an external MCP server the user has not authorized.
  It points the user to the external servers popover to connect; afterwards they send their message again.
  The card stays in the history and only changes its text once the server is connected.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlugZap } from "lucide-svelte";
  import { getMcpConnectContext } from "../../../connect/mcp-connect-context";
  import {
    createMcpServersQuery,
    openMcpServersPopover,
  } from "../../../connect/mcp-servers";
  import type { McpAuthRequiredBlock } from "./mcp-auth-required-block";

  export let block: McpAuthRequiredBlock;

  const runtimeClient = useRuntimeClient();
  const connectContext = getMcpConnectContext();
  const serversQuery = createMcpServersQuery(runtimeClient);

  $: server = ($serversQuery.data ?? []).find(
    (s) => s.connector === block.connector,
  );
  $: connected = !!server?.connected;
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
      {#if connected}
        {m.chat_mcp_auth_connected_body({ name: block.displayName })}
      {:else if connectContext}
        {m.chat_mcp_auth_required_body({ name: block.displayName })}
      {:else}
        {m.chat_mcp_auth_required_no_provider({ name: block.displayName })}
      {/if}
    </p>
    {#if !connected && connectContext}
      <div class="mcp-auth-actions">
        <Button type="primary" small onClick={openMcpServersPopover}>
          {m.chat_mcp_connect_server({ name: block.displayName })}
        </Button>
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

  .mcp-auth-actions {
    @apply mt-1 flex gap-x-2;
  }
</style>
