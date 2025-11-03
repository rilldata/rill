<script lang="ts">
  import type { PageData } from "./$types";
  import ScheduledReportEditor from "@rilldata/web-common/features/scheduled-reports/ScheduledReportEditor.svelte";

  export let data: PageData;
  const { report, organization, project, reportName } = data;
</script>

<div class="flex flex-row size-full max-w-full max-h-full overflow-auto">
  <div class="w-full overflow-auto">
    <slot />
  </div>
  <ScheduledReportEditor
    {organization}
    {project}
    props={{
      mode: "edit",
      reportName,
      reportSpec: report.report?.spec ?? {},
    }}
  />
</div>
