<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectDeploymentStatus from "@rilldata/web-admin/features/projects/status/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/status/ProjectGithubConnection.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex flex-col gap-y-12">
      <div class="flex gap-x-20 items-start">
        <ProjectGithubConnection {organization} {project} />
        <ProjectDeploymentStatus {organization} {project} />
      </div>
      <ProjectResources />
      <ProjectParseErrors />
    </div>
  </ContentContainer>
</VerticalScrollContainer>
