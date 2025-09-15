<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import ProjectSelector from "@rilldata/web-common/features/project/deploy/ProjectSelector.svelte";
  import ProjectSelectorItem from "@rilldata/web-common/features/project/deploy/ProjectSelectorItem.svelte";
  import RequestProjectAccessDialog from "@rilldata/web-common/features/project/deploy/RequestProjectAccessDialog.svelte";
  import { getUpdateProjectRoute } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { getManageProjectAccess } from "@rilldata/web-common/features/project/selectors.ts";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import Rocket from "svelte-radix/Rocket.svelte";

  export let open = false;
  export let matchingProjects: Project[];

  let selectedProject: Project | undefined =
    matchingProjects.length === 1 ? matchingProjects[0] : undefined;
  $: hasManageProjectAccess = getManageProjectAccess(
    selectedProject?.orgName ?? "",
    selectedProject?.name ?? "",
  );

  $: enableUpdate = !!selectedProject;

  $: deployUrl = selectedProject
    ? getUpdateProjectRoute(selectedProject.orgName, selectedProject.name)
    : "";

  let showRequestProjectAccess = false;
</script>

<Popover.Root bind:open>
  <Popover.Trigger asChild let:builder>
    <Button type="primary" builders={[builder]}>
      <Rocket size="16px" />

      Deploy
    </Button>
  </Popover.Trigger>
  <Popover.Content align="start" class="w-[420px] flex flex-col gap-y-2">
    <div class="text-base font-medium">Update</div>
    <div class="text-sm text-slate-500">Push local changes to Rill Cloud?</div>

    {#if matchingProjects.length === 1 && selectedProject}
      <div class="border rounded-sm border-gray-300">
        <ProjectSelectorItem project={selectedProject} />
      </div>
    {:else}
      <ProjectSelector bind:selectedProject projects={matchingProjects} />
    {/if}

    <div class="flex flex-row-reverse items-center gap-x-2">
      {#if $hasManageProjectAccess || !selectedProject}
        <Button
          type="primary"
          disabled={!enableUpdate}
          href={deployUrl}
          target="_blank"
          onClick={() => (open = false)}
        >
          Update
        </Button>
      {:else}
        <Button
          type="primary"
          disabled={!enableUpdate}
          onClick={() => {
            open = false;
            showRequestProjectAccess = true;
          }}
        >
          Update
        </Button>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

{#if selectedProject}
  <RequestProjectAccessDialog
    bind:open={showRequestProjectAccess}
    project={selectedProject}
  />
{/if}
