<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import { getManageProjectAccess } from "@rilldata/web-common/features/project/selectors.ts";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";
  import { createLocalServiceListProjectsForOrgRequest } from "@rilldata/web-common/runtime-client/local-service.ts";
  import ProjectSelector from "@rilldata/web-common/features/project/deploy/ProjectSelector.svelte";
  import RequestProjectAccessDialog from "@rilldata/web-common/features/project/deploy/RequestProjectAccessDialog.svelte";
  import OverwriteProjectConfirmationDialog from "@rilldata/web-common/features/project/deploy/OverwriteProjectConfirmationDialog.svelte";

  $: orgParam = $page.url.searchParams.get("org") ?? "";

  $: projectsForOrg = createLocalServiceListProjectsForOrgRequest(orgParam, {
    query: {
      enabled: !!orgParam,
    },
  });

  let selectedProject: Project | undefined = undefined;
  $: hasManageProjectAccess = getManageProjectAccess(
    selectedProject?.orgName ?? "",
    selectedProject?.name ?? "",
  );
  $: deployUrl = `/deploy/update?org=${selectedProject?.orgName}&project=${selectedProject?.name}&new_managed_repo=true`;

  let showOverwriteProjectConfirmation = false;
  let showRequestProjectAccess = false;

  function onUpdateProject() {
    if ($hasManageProjectAccess) {
      showOverwriteProjectConfirmation = true;
    } else {
      showRequestProjectAccess = true;
    }
  }
</script>

<div class="flex flex-col gap-y-2">
  <div class="text-xl">Which project would you like to overwrite?</div>
  <div class="text-sm text-gray-500">
    These are all the projects listed under the selected org <b>{orgParam}</b>.
  </div>
  <div class="w-[500px]">
    <ProjectSelector
      bind:selectedProject
      projects={$projectsForOrg.data?.projects}
    />
  </div>
</div>

<Button
  wide
  type="primary"
  onClick={onUpdateProject}
  disabled={!selectedProject}
>
  Update selected project
</Button>
<Button wide type="ghost" on:click={() => history.back()}>Back</Button>

<OverwriteProjectConfirmationDialog
  bind:open={showOverwriteProjectConfirmation}
  {deployUrl}
  rillManagedProject={!!selectedProject?.managedGitId}
/>

{#if selectedProject}
  <RequestProjectAccessDialog
    bind:open={showRequestProjectAccess}
    project={selectedProject}
  />
{/if}
