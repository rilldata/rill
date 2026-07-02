<!-- 
  This layout wraps the chat page to provide proper height constraints
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import MCPConnectDialog from "@rilldata/web-admin/features/ai/mcp/MCPConnectDialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: projectQuery = createAdminServiceGetProject(organization, project);
  $: isPublic = $projectQuery.data?.project?.public ?? true;

  let mcpDialogOpen = false;
</script>

<div class="chat-page-wrapper">
  <ProjectChat>
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
