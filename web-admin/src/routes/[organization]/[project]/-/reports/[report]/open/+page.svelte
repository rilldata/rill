<script lang="ts">
  import ReportOpen from "@rilldata/web-admin/features/scheduled-reports/open/ReportOpen.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
</script>

<ReportOpen
  reportSpec={data.reportSpec}
  exploreName={data.exploreName}
  executionTime={data.executionTime}
  pathContext={{
    organization: data.organization,
    project: data.project,
    token: data.token,
  }}
  ignoreFilters={!!data.token}
  backHref={`/${data.organization}/${data.project}/-/reports/${data.reportId}`}
/>
