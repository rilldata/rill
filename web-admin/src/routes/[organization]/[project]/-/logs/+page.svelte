<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentLogs from "@rilldata/web-admin/features/projects/ProjectDeploymentLogs.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ProjectDeploymentStatus from "../../../../../features/projects/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "./../../../../../features/projects/ProjectGithubConnection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <div class="px-9 pt-4 pb-6 flex gap-x-16 items-start">
    <ProjectDeploymentStatus {organization} {project} />
    <ProjectGithubConnection {organization} {project} />
  </div>

  <ProjectDeploymentLogs {organization} {project} />
</VerticalScrollContainer>
