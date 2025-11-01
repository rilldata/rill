<script lang="ts">
  import ScheduledReportEditor from "@rilldata/web-common/features/scheduled-reports/ScheduledReportEditor.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { organization, project, exploreName } = data;
</script>

<div class="flex flex-row w-screen h-full">
  <div class="w-full">
    <slot />
  </div>
  <ScheduledReportEditor
    {organization}
    {project}
    props={{ mode: "create", exploreName }}
  />
</div>
