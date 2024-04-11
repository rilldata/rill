<script lang="ts">
  import AlertBreadcrumb from "./AlertBreadcrumb.svelte";
  import DashboardBreadcrumb from "./DashboardBreadcrumb.svelte";
  import OrganizationBreadcrumb from "./OrganizationBreadcrumb.svelte";
  import ProjectBreadcrumb from "./ProjectBreadcrumb.svelte";
  import ReportBreadcrumb from "./ReportBreadcrumb.svelte";
</script>

<nav>
  <ol class="flex flex-row items-center">
    <OrganizationBreadcrumb />
    <ProjectBreadcrumb />
    <DashboardBreadcrumb />
    <ReportBreadcrumb />
    <AlertBreadcrumb />
  </ol>
</nav>
