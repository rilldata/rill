<script lang="ts">
  import { page } from "$app/stores";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

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
</ProjectPage>
