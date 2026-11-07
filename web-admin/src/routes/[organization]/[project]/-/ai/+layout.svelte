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

  $: onPublicURLPage = isPublicURLPage($page);
  const user = createAdminServiceGetCurrentUser();
  $: loggedIn = !!$user.data?.user;

  // Cookie-authenticated project query. On a public URL page (a shared conversation opened with a magic token) it tells us
  // whether the logged-in visitor has access to the project in their own right; anonymous visitors can't run it.
  $: projectQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    { query: { enabled: loggedIn } },
  );
  $: isPublic = $projectQuery.data?.project?.public ?? true;

  // Visitors who can't continue the conversation get a read-only view: anonymous visitors, and logged-in visitors of a shared
  // conversation who don't have access to the project (continuing forks the conversation under their own credentials).
  $: readOnly = !loggedIn || (onPublicURLPage && !$projectQuery.isSuccess);

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
  <ProjectChat {readOnly} beforeFork={switchToUserCredentialsBeforeFork}>
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
