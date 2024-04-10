<script lang="ts">
  import AlertBreadcrumb from "./AlertBreadcrumb.svelte";
  import OrganizationBreadcrumb from "./OrganizationBreadcrumb.svelte";
  import ProjectBreadcrumb from "./ProjectBreadcrumb.svelte";
  import ReportBreadcrumb from "./ReportBreadcrumb.svelte";
  import VisualizationBreadcrumb from "./VisualizationBreadcrumb.svelte";
</script>

<nav>
  <ol class="flex flex-row items-center">
    <OrganizationBreadcrumb />
    <ProjectBreadcrumb />
    <VisualizationBreadcrumb />
    <ReportBreadcrumb />
    <AlertBreadcrumb />
  </ol>
</nav>
