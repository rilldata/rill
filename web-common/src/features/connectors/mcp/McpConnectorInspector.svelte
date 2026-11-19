<!--
  Side panel for connector files with `driver: mcp`.
  Shows the server's status and the tools it exposes to the current user.
  Connecting and disconnecting OAuth servers happens in the chat's external servers popover.
-->
<script lang="ts">
  import { createMcpServersQuery } from "@rilldata/web-common/features/chat/connect/mcp-servers";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { CheckCircle2, AlertCircle, Plug, PlugZap } from "lucide-svelte";

  export let resource: V1Resource | undefined;
  export let connectorName: string;

  const runtimeClient = useRuntimeClient();
  const serversQuery = createMcpServersQuery(runtimeClient);

  $: spec = resource?.connector?.spec;
  $: properties = (spec?.properties ?? {}) as Record<string, unknown>;
  $: displayName = (properties.display_name as string) || connectorName;
  $: url = properties.url as string | undefined;
  $: oauth = properties.auth === "oauth";
  $: reconcileError = resource?.meta?.reconcileError;
  $: server = ($serversQuery.data ?? []).find(
    (s) => s.connector === connectorName,
  );
  $: connected = !!server?.connected;
  $: tools = server?.tools ?? [];
  $: agents = spec?.agents ?? [];
</script>

<div class="mcp-inspector">
  <div class="mcp-section">
    <div class="mcp-heading">
      <Plug size="14px" />
      <span>{m.mcp_connector_panel_title()}</span>
    </div>
    <div class="mcp-name">{displayName}</div>
    {#if url}
      <div class="mcp-url" title={url}>{url}</div>
    {/if}
    <div class="mcp-hint">
      {oauth
        ? m.mcp_connector_panel_auth_oauth()
        : m.mcp_connector_panel_auth_static()}
    </div>
  </div>

  <div class="mcp-section">
    <div class="mcp-label">{m.mcp_connector_panel_status()}</div>
    {#if reconcileError}
      <div class="mcp-status error">
        <AlertCircle size="14px" />
        <span>{m.mcp_connector_panel_reconcile_error()}</span>
      </div>
      <pre class="mcp-error">{reconcileError}</pre>
    {:else if server?.error}
      <div class="mcp-status error">
        <AlertCircle size="14px" />
        <span>{m.chat_mcp_status_error()}</span>
      </div>
      <pre class="mcp-error">{server.error}</pre>
    {:else if $serversQuery.isLoading}
      <div class="mcp-status">
        <span>{m.common_loading()}</span>
      </div>
    {:else if !connected}
      <div class="mcp-status" class:error={!oauth}>
        {#if oauth}
          <PlugZap size="14px" />
          <span>{m.chat_mcp_status_not_connected()}</span>
        {:else}
          <AlertCircle size="14px" />
          <span>{m.chat_mcp_status_error()}</span>
        {/if}
      </div>
    {:else}
      <div class="mcp-status connected">
        <CheckCircle2 size="14px" />
        <span>{m.chat_mcp_status_connected()}</span>
      </div>
    {/if}
  </div>

  <div class="mcp-section">
    <div class="mcp-label">{m.mcp_connector_panel_agents()}</div>
    <div class="mcp-hint">
      {agents.length ? agents.join(", ") : m.mcp_connector_panel_all_agents()}
    </div>
  </div>

  <div class="mcp-section">
    <div class="mcp-label">
      {m.mcp_connector_panel_tools()}
      {#if tools.length}
        <span class="mcp-count">{tools.length}</span>
      {/if}
    </div>
    {#if tools.length}
      <ul class="mcp-tools">
        {#each tools as tool (tool.name)}
          <li class="mcp-tool">
            <div class="mcp-tool-name">{tool.name}</div>
            {#if tool.description}
              <div class="mcp-tool-description">{tool.description}</div>
            {/if}
          </li>
        {/each}
      </ul>
    {:else if oauth && !connected}
      <div class="mcp-hint">{m.mcp_connector_panel_tools_per_user()}</div>
    {:else}
      <div class="mcp-hint">{m.mcp_connector_panel_no_tools()}</div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .mcp-inspector {
    @apply flex h-full w-full flex-col gap-y-4 overflow-y-auto p-4;
    @apply border-l bg-surface-background;
  }

  .mcp-section {
    @apply flex flex-col gap-y-1;
  }

  .mcp-heading {
    @apply flex items-center gap-x-1.5 text-xs font-semibold uppercase text-fg-muted;
  }

  .mcp-name {
    @apply text-sm font-medium text-fg-primary;
  }

  .mcp-url {
    @apply truncate text-xs text-fg-secondary;
  }

  .mcp-label {
    @apply flex items-center gap-x-1.5 text-xs font-semibold text-fg-secondary;
  }

  .mcp-count {
    @apply rounded bg-surface-subtle px-1.5 text-xs font-normal text-fg-muted;
  }

  .mcp-hint {
    @apply text-xs text-fg-muted;
  }

  .mcp-status {
    @apply flex items-center gap-x-1.5 text-sm text-fg-secondary;
  }

  .mcp-status.connected {
    @apply text-fg-primary;
  }

  .mcp-status.error {
    @apply text-red-600;
  }

  .mcp-error {
    @apply whitespace-pre-wrap break-words rounded bg-surface-subtle p-2 text-xs text-red-600;
  }

  .mcp-tools {
    @apply flex flex-col gap-y-2;
  }

  .mcp-tool-name {
    @apply font-mono text-xs text-fg-primary;
  }

  .mcp-tool-description {
    @apply text-xs text-fg-muted;
  }
</style>
