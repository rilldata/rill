<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectDeploymentLogs from "@rilldata/web-admin/features/projects/ProjectDeploymentLogs.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ProjectDeploymentStatus from "../../../../../features/projects/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "./../../../../../features/projects/ProjectGithubConnection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <!-- hero -->
    <div class="flex gap-x-16 mb-12 items-start">
      <ProjectDeploymentStatus {organization} {project} />
      <ProjectGithubConnection {organization} {project} />
    </div>

    <!-- contained logs -->
    <ProjectDeploymentLogs {organization} {project} />
  </ContentContainer>
</VerticalScrollContainer>
