<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import { createLocalServiceListProjectsForOrgRequest } from "@rilldata/web-common/runtime-client/local-service";
  import ProjectSelector from "@rilldata/web-common/features/project/ProjectSelector.svelte";

  $: orgParam = $page.url.searchParams.get("org") ?? "";

  $: projectsForOrg = createLocalServiceListProjectsForOrgRequest(orgParam, {
    query: {
      enabled: !!orgParam,
    },
  });

  let selectedProject: Project | undefined = undefined;
</script>

<div class="text-xl">Which project would you like to overwrite?</div>
<div class="text-base text-gray-500">
  These are all the projects listed under the selected org <b>{orgParam}</b>.
</div>
<ProjectSelector
  bind:selectedProject
  projects={$projectsForOrg.data?.projects}
/>

<Button
  wide
  type="primary"
  href="/deploy/redeploy?org={selectedProject?.orgName}&project_id={selectedProject?.id}"
  disabled={!selectedProject}
>
  Update selected project
</Button>
