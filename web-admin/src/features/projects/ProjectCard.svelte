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
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { readable } from "svelte/store";

  let { organization, project }: { organization: string; project: string } =
    $props();

  // The card has room for this many tags; the rest are summarized in a "+N" chip.
  const VISIBLE_TAGS = 2;

  // Check whether project is public or private
  let proj = $derived(createAdminServiceGetProject(organization, project));
  let primaryBranch = $derived($proj.data?.project?.primaryBranch);
  let tags = $derived($proj.data?.project?.tags ?? []);
  let hiddenTags = $derived(tags.slice(VISIBLE_TAGS));

  let hovering = $state(false);
  let actionsOpen = $state(false);
  let editProjectOpen = $state(false);
  let renameProjectOpen = $state(false);
  let deleteProjectOpen = $state(false);
  const disabledCloudEditingQuery = readable({ data: false });

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

  let cloudEditingQuery = $derived.by(() => {
    if (!runtimeConfig) return disabledCloudEditingQuery;
    const runtimeClient = getRuntimeClient(runtimeConfig);
    return createRuntimeServiceGetInstance(
      runtimeClient,
      {},
      {
        query: {
          select: (data) => !!data.instance?.featureFlags?.cloudEditing,
        },
      },
    );
  });

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
        <svelte:fragment slot="read-project"
          >{m.project_role_viewer()}</svelte:fragment
        >
        <svelte:fragment slot="manage-project"
          >{m.project_role_admin()}</svelte:fragment
        >
      </ProjectAccessControls>
    </Tag>
    <!-- Project tags -->
    {#if tags.length > 0}
      <ul class="tags" aria-label={m.project_card_tags_label()}>
        {#each tags.slice(0, VISIBLE_TAGS) as tag (tag)}
          <li class="tag" title={tag}>{tag}</li>
        {/each}
        {#if hiddenTags.length > 0}
          <li class="tag shrink-0" title={hiddenTags.join(", ")}>
            {m.project_card_tags_more({ count: hiddenTags.length })}
          </li>
        {/if}
      </ul>
    {/if}
    <!-- Public vs Private indicator -->
    <div class="absolute bottom-2.5 right-2.5 text-fg-secondary">
      <Tooltip distance={10}>
        {#if projectData.public}
          <Globe size="16px" />
        {:else}
          <Lock size="16px" />
        {/if}
        <TooltipContent slot="tooltip-content">
          <span class="text-xs">
            {projectData.public
              ? m.project_card_visibility_public()
              : m.project_card_visibility_private()}
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

<style lang="postcss">
  .tags {
    @apply absolute bottom-2.5 left-2.5 right-10;
    @apply flex flex-row items-center gap-x-1 overflow-hidden;
  }

  .tag {
    @apply h-[21px] max-w-[76px] truncate px-2 rounded-[20px] border;
    @apply bg-gray-100 border-gray-300 text-fg-secondary text-xs leading-[19px];
  }

  :global(.dark) .tag {
    @apply border-gray-400 bg-surface-active;
  }
</style>
