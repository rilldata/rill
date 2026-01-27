<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectClone from "@rilldata/web-admin/features/projects/status/ProjectClone.svelte";
  import ProjectDeploymentStatus from "@rilldata/web-admin/features/projects/status/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ContentContainer maxWidth={1100} showTitle={false}>
  <div class="flex flex-col gap-y-4 size-full">
    <h1 class="text-2xl font-semibold" aria-label="Container title">
      Project status
    </h1>

    <div class="flex justify-between items-start">
      <div class="flex gap-x-20 items-start">
        <ProjectGithubConnection {organization} {project} />
        <ProjectDeploymentStatus {organization} {project} />
      </div>
      <ProjectClone {organization} {project} />
    </div>

    <ProjectResources />
    <ProjectParseErrors />
  </div>
</ContentContainer>
