<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import { createLocalServiceListMatchingProjectsRequest } from "@rilldata/web-common/runtime-client/local-service";
  import ProjectSelector from "@rilldata/web-common/features/project/deploy/ProjectSelector.svelte";

  const matchingProjects = createLocalServiceListMatchingProjectsRequest();

  let selectedProject: Project | undefined = undefined;
</script>

<div class="flex flex-col gap-y-2">
  <div class="text-xl">Which project would you like to update?</div>
  <div class="text-sm text-gray-500">
    These all have matching project name with your Rill Developer project.
  </div>
  <div class="w-[500px]">
    <ProjectSelector
      bind:selectedProject
      projects={$matchingProjects.data?.projects}
    />
  </div>
</div>

<Button
  wide
  type="primary"
  href="/deploy/redeploy?org={selectedProject?.orgName}&project={selectedProject?.name}"
  disabled={!selectedProject}
>
  Update selected project
</Button>
<Button wide type="ghost" href="/deploy/select-org">
  Or deploy to other project
</Button>
