<script lang="ts">
  import Card from "@rilldata/web-common/components/card/Card.svelte";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetProject } from "../../client";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";
  import ProjectCardActions from "@rilldata/web-admin/features/projects/ProjectCardActions.svelte";
  import GuardedDeleteProjectConfirmation from "@rilldata/web-admin/features/projects/settings/GuardedDeleteProjectConfirmation.svelte";
  import ProjectRenameDialog from "@rilldata/web-admin/features/projects/settings/ProjectRenameDialog.svelte";
  import EditBranchDialog from "@rilldata/web-admin/features/edit-session/EditBranchDialog.svelte";
  import { getFeatureFlags } from "@rilldata/web-common/features/feature-flags";
  import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery } from "@tanstack/svelte-query";

  let { organization, project }: { organization: string; project: string } =
    $props();

  // Check whether project is public or private
  let proj = $derived(createAdminServiceGetProject(organization, project));
  let primaryBranch = $derived($proj.data?.project?.primaryBranch);

  let hovering = $state(false);
  let actionsOpen = $state(false);
  let editProjectOpen = $state(false);
  let renameProjectOpen = $state(false);
  let deleteProjectOpen = $state(false);

  function doesProjectNameIncludeUnderscores(project: string) {
    return project.includes("_");
  }

  let runtimeConfig = $derived.by(() => {
    const deployment = $proj.data?.deployment;
    if (!deployment?.runtimeHost || !deployment.runtimeInstanceId) {
      return undefined;
    }

    return {
      host: deployment.runtimeHost,
      instanceId: deployment.runtimeInstanceId,
      jwt: $proj.data?.jwt,
    };
  });

  let cloudEditingQuery = $derived(
    createQuery({
      queryKey: [
        "project-card-cloud-editing",
        organization,
        project,
        runtimeConfig?.host,
        runtimeConfig?.instanceId,
      ],
      enabled: !!runtimeConfig,
      queryFn: async () => {
        if (!runtimeConfig) return false;
        const flags = await getFeatureFlags(getRuntimeClient(runtimeConfig));
        return !!flags.cloudEditing;
      },
    }),
  );

  let canEditProject = $derived(
    !!$cloudEditingQuery.data && !!$proj.data?.projectPermissions?.manageDev,
  );

  $effect(() => {
    if (!canEditProject) editProjectOpen = false;
  });
</script>

{#if $proj.data?.project}
  {@const projectData = $proj.data.project}
  <Card href="/{organization}/{project}" bind:hovering>
    <!-- Project name -->
    <h2
      class="text-fg-primary font-medium text-lg text-center px-4 {doesProjectNameIncludeUnderscores(
        project,
      )
        ? 'break-all'
        : 'break-words'}"
    >
      {project}
    </h2>
    <!-- Project actions -->
    {#if hovering || actionsOpen}
      <div class="absolute top-2.5 right-2.5 text-fg-secondary">
        <ProjectCardActions
          {organization}
          {project}
          bind:open={actionsOpen}
          canEdit={canEditProject}
          onEdit={() => (editProjectOpen = true)}
          onRename={() => (renameProjectOpen = true)}
          onDelete={() => (deleteProjectOpen = true)}
        />
      </div>
    {/if}
    <!-- Permissions tag -->
    <Tag>
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="read-project">Viewer</svelte:fragment>
        <svelte:fragment slot="manage-project">Admin</svelte:fragment>
      </ProjectAccessControls>
    </Tag>
    <!-- Public vs Private indicator -->
    <div class="absolute bottom-2.5 right-2.5 text-fg-secondary">
      <Tooltip distance={10}>
        {#if projectData.public}
          <Globe size="16px" />
        {:else}
          <Lock size="16px" />
        {/if}
        <TooltipContent slot="tooltip-content">
          <span class="text-xs"
            >This project is
            {#if projectData.public}
              <span class="font-medium"> public</span>
            {:else}
              <span class="font-medium"> private</span>
            {/if}
          </span>
        </TooltipContent>
      </Tooltip>
    </div>
  </Card>
{/if}

<ProjectRenameDialog {organization} {project} bind:open={renameProjectOpen} />

<GuardedDeleteProjectConfirmation
  {organization}
  {project}
  bind:open={deleteProjectOpen}
  button={false}
/>

{#if canEditProject}
  <EditBranchDialog
    bind:open={editProjectOpen}
    {organization}
    {project}
    {primaryBranch}
  />
{/if}
