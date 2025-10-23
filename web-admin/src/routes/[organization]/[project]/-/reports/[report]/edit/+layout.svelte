<script lang="ts">
  import type { PageData } from "./$types";
  import ScheduledReportEditor from "@rilldata/web-common/features/scheduled-reports/ScheduledReportEditor.svelte";

  export let data: PageData;
  const { report, organization, project, reportName } = data;
</script>

<div class="flex flex-row w-screen">
  <div class="w-full">
    <slot />
  </div>
  <div class="w-[500px]">
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
</div>
