<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { parseReport } from "@rilldata/web-admin/features/scheduled-reports/query-mapper";
  import { useReport } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportId = $page.params.report;
  $: executionTime = $page.url.searchParams.get("execution_time");

  $: report = useReport($runtime.instanceId, reportId);

  $: parsedReport = parseReport($report.data?.resource, executionTime);

  $: if ($parsedReport.ready) {
    goto(
      `/${organization}/${project}/${$parsedReport.metricsView}?state=${$parsedReport.state}`,
    );
  }

  // TODO: error handling
</script>
