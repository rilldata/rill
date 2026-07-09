<!--
  Wires the shared connect-client context for Rill Cloud and renders the
  MCPConnectDialog once for all descendant chat surfaces (global, dashboard,
  developer). Rill Developer and embedded dashboards never mount this, so their
  chat surfaces have no provider and render no connect CTA.
-->
<script lang="ts">
  import { setConnectClientContext } from "@rilldata/web-common/features/chat/connect/connect-client-context";
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

  setConnectClientContext({ open: () => (dialogOpen = true) });
</script>

{@render children()}

<MCPConnectDialog bind:open={dialogOpen} {organization} {project} {isPublic} />
