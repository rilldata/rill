<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentStatus from "@rilldata/web-admin/features/projects/status/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <div class="flex gap-x-20 items-start">
    <ProjectGithubConnection {organization} {project} />
    <ProjectDeploymentStatus {organization} {project} />
  </div>

  <ProjectResources />
  <ProjectParseErrors />
</div>
