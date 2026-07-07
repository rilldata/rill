<!--
  Wires the shared connect-client context for Rill Cloud and renders the
  MCPConnectDialog once for all descendant chat surfaces (global, dashboard,
  developer). Rill Developer never mounts this, so its chat surfaces leave the
  context disabled and show no connect CTA.
-->
<script lang="ts">
  import {
    setConnectClientContext,
    type ConnectClientProvider,
  } from "@rilldata/web-common/features/chat/connect/connect-client-context";
  import type { Snippet } from "svelte";
  import MCPConnectDialog from "./MCPConnectDialog.svelte";

  let {
    organization,
    project,
    isPublic,
    children,
  }: {
    organization: string;
    project: string;
    isPublic: boolean;
    children: Snippet;
  } = $props();

  let dialogOpen = $state(false);
  let selectedProvider = $state<ConnectClientProvider | null>(null);

  setConnectClientContext({
    enabled: true,
    open: (provider) => {
      selectedProvider = provider ?? null;
      dialogOpen = true;
    },
  });
</script>

{@render children()}

<MCPConnectDialog
  bind:open={dialogOpen}
  {organization}
  {project}
  {isPublic}
  provider={selectedProvider}
/>
