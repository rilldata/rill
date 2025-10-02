<!-- 
  This layout wraps the chat page to provide proper height constraints
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import MCPConfigDialog from "@rilldata/web-admin/features/ai/MCPConfigDialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import ProjectChat from "@rilldata/web-common/features/chat/ProjectChat.svelte";

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
        gray
        onClick={() => (mcpDialogOpen = true)}
        class="w-full"
      >
        <APIIcon size="14px" />
        Connect your own client
      </Button>
    </svelte:fragment>
  </ProjectChat>

  <MCPConfigDialog
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
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: #ffffff;
    min-height: 0;
  }
</style>
