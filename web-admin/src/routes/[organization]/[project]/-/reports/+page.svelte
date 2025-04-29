<script lang="ts">
  import { page } from "$app/stores";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";

  $: ({ instanceId } = $runtime);

  $: ({
    params: { organization, project },
  } = $page);

  $: query = useReports(instanceId);

  $: ({ data } = $query);

  $: reports = data?.resources ?? [];
</script>

<ProjectPage {query} kind="report">
  <ReportsTable {organization} {project} data={reports} slot="table" />
  <svelte:fragment slot="action">
    To create a report, click the Export button in a dashboard.
  </svelte:fragment>
</ProjectPage>
