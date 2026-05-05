<script lang="ts">
  import { page } from "$app/stores";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  $: ({
    params: { organization, project },
  } = $page);

  $: query = useReports(runtimeClient);

  $: ({ data } = $query);

  $: reports = data?.resources ?? [];
</script>

<ProjectPage {query} kind="report">
  <ReportsTable {organization} {project} data={reports} slot="table" />
</ProjectPage>
