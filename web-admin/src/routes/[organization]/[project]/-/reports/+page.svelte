<script lang="ts">
  import { page } from "$app/stores";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";

  $: ({ instanceId } = $runtime);

  const { fullPageReportEditor } = featureFlags;

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
    {#if $fullPageReportEditor}
      <Button
        type="link"
        href="/{organization}/{project}/-/reports/-/create"
        large
      >
        Create a new report
      </Button>
    {:else}
      To create a report, click the Export button in a dashboard.
    {/if}
  </svelte:fragment>
</ProjectPage>
