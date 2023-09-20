<script lang="ts">
  import { page } from "$app/stores";
  import ProjectDeploymentLogs from "@rilldata/web-admin/features/projects/ProjectDeploymentLogs.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <ProjectDeploymentLogs {organization} {project} />
</VerticalScrollContainer>
