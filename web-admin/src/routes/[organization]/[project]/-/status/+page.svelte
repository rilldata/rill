<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentStatus from "@rilldata/web-admin/features/projects/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/ProjectGithubConnection.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/ProjectParseErrors.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <div class="pt-4 flex flex-col gap-y-6">
    <div class="px-12 flex gap-x-20 items-start">
      <ProjectDeploymentStatus {organization} {project} />
      <ProjectGithubConnection {organization} {project} />
    </div>
    <ProjectParseErrors {organization} {project} />
  </div>
</VerticalScrollContainer>
