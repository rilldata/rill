<!--
  Sets the connect-client context so ConnectClientPopover instances in descendant
  chat surfaces can open the MCPConnectDialog rendered here. Chat surfaces without
  this provider (Rill Developer, embeds, edit mode) render no connect CTA.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { setConnectClientContext } from "@rilldata/web-common/features/chat/connect/connect-client-context";
  import MCPConnectDialog from "./MCPConnectDialog.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: projectQuery = createAdminServiceGetProject(organization, project);
  $: isPublic = $projectQuery.data?.project?.public ?? true;

  let dialogOpen = false;

  setConnectClientContext({ open: () => (dialogOpen = true) });
</script>

<slot />

<MCPConnectDialog bind:open={dialogOpen} {organization} {project} {isPublic} />
