<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import ProjectOverview from "@rilldata/web-admin/features/projects/status/overview/ProjectOverview.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ContentContainer title="Project Status" maxWidth={1100}>
  <ProjectOverview {organization} {project} />
</ContentContainer>
