<!-- 
  This layout wraps the chat page to provide proper height constraints
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
  } from "@rilldata/web-admin/client";
  import MCPConnectDialog from "@rilldata/web-admin/features/ai/mcp/MCPConnectDialog.svelte";
  import { isPublicURLPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";
  import { setConnectClientContext } from "@rilldata/web-common/features/chat/connect/connect-client-context";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { fetchProjectDeploymentDetails } from "@rilldata/web-admin/features/projects/selectors";
  import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  // Anonymous visitors of a shared conversation (public URL page) are not logged in,
  // so the cookie-authenticated project query would fail; it's only needed for the MCP dialog.
  $: onPublicURLPage = isPublicURLPage($page);
  $: projectQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    { query: { enabled: !onPublicURLPage } },
  );
  $: isPublic = $projectQuery.data?.project?.public ?? true;

  // Anonymous visitors (e.g. recipients of an AI report opened with a magic token) get a read-only view of the conversation.
  const user = createAdminServiceGetCurrentUser();
  $: loggedIn = !!$user.data?.user;

  // A logged-in user opening a shared conversation with a magic token reads it through the token,
  // but the token only grants read access to the conversation. Continuing it forks the conversation,
  // which must run with the user's own credentials to have access to the project's data.
  // So right before forking, fetch the user's runtime JWT with their cookie session and put it on the shared runtime client
  // (the one the conversation manager uses), then drop the token from the URL so the layouts also re-authenticate as the user.
  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  async function switchToUserCredentialsBeforeFork() {
    if (!$page.url.searchParams.has("token")) return;

    const { runtime } = await fetchProjectDeploymentDetails(
      organization,
      project,
      undefined,
    );
    if (!runtime.jwt) {
      throw new Error("No runtime credentials for the current user");
    }
    runtimeClient.updateJwt(runtime.jwt.token, "user");
    // Data fetched through the token was restricted to what the token could access, so refetch it as the user.
    await invalidateRuntimeQueries(queryClient, runtimeClient.instanceId);

    const target = new URL($page.url);
    target.searchParams.delete("token");
    await goto(target, { replaceState: true, noScroll: true, keepFocus: true });
    // The layouts now provide the user's JWT as well; re-apply ours in case a stale token JWT was pushed during the transition.
    runtimeClient.updateJwt(runtime.jwt.token, "user");
  }

  let mcpDialogOpen = false;

  // Lets the ConnectClientPopover in the chat header open the MCPConnectDialog below.
  setConnectClientContext({ open: () => (mcpDialogOpen = true) });
</script>

<div class="chat-page-wrapper">
  <ProjectChat
    readOnly={!loggedIn}
    beforeFork={switchToUserCredentialsBeforeFork}
  >
    <svelte:fragment slot="sidebar-footer">
      <Button
        type="secondary"
        onClick={() => (mcpDialogOpen = true)}
        class="w-full"
      >
        <APIIcon size="14px" className="!fill-current" />
        {m.chat_connect_client()}
      </Button>
    </svelte:fragment>
    <svelte:fragment slot="sidebar-collapsed-footer">
      <Button
        type="secondary"
        square
        label={m.chat_connect_client()}
        onClick={() => (mcpDialogOpen = true)}
      >
        <APIIcon size="14px" className="!fill-current" />
      </Button>
    </svelte:fragment>
  </ProjectChat>

  <MCPConnectDialog
    bind:open={mcpDialogOpen}
    {organization}
    {project}
    {isPublic}
  />

  <!-- This slot isn't used, but its presence avoids a SvelteKit browser console warning. -->
  <slot />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    @apply bg-surface-background;
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;

    min-height: 0;
  }
</style>
