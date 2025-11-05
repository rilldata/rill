<script lang="ts">
  import type { PageData } from "./$types";
  import ScheduledReportEditor from "@rilldata/web-common/features/scheduled-reports/ScheduledReportEditor.svelte";

  export let data: PageData;
  const { report, organization, project, reportName, aggregationRequest } =
    data;
</script>

<ScheduledReportEditor
  {organization}
  {project}
  props={{
    mode: "edit",
    reportName,
    reportSpec: report.report?.spec ?? {},
  }}
  {aggregationRequest}
/>
