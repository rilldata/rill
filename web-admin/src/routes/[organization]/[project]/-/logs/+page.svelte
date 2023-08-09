<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentLogs from "@rilldata/web-admin/components/projects/ProjectDeploymentLogs.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ProjectDeploymentLogs {organization} {project} />
