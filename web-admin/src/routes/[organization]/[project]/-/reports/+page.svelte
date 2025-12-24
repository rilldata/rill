<script lang="ts">
  import { page } from "$app/stores";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";

  const instanceId = httpClient.getInstanceId();

  $: ({
    params: { organization, project },
  } = $page);

  $: query = useReports(instanceId);

  $: ({ data } = $query);

  $: reports = data?.resources ?? [];
</script>

<ProjectPage {query} kind="report">
  <ReportsTable {organization} {project} data={reports} slot="table" />
</ProjectPage>
