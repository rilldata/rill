<script lang="ts">
  import AlertOpen from "@rilldata/web-admin/features/alerts/open/AlertOpen.svelte";
  import { getCloudDashboardPaths } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
</script>

<AlertOpen
  alertSpec={data.alertSpec}
  exploreName={data.exploreName}
  executionTime={data.executionTime}
  dashboardPaths={getCloudDashboardPaths(
    data.organization,
    data.project,
    data.token,
  )}
  ignoreFilters={!!data.token}
  backHref={`/${data.organization}/${data.project}/-/alerts/${data.alertId}`}
/>
