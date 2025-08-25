<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import {
    getDashboardToRedirect,
    useRefetchingDashboards,
  } from "@rilldata/web-admin/features/dashboards/listing/refetching-dashboards.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let pageData: PageData;
  const { deploying, deployingName } = pageData;

  $: ({
    params: { organization, project },
  } = $page);
  $: ({ instanceId } = $runtime);

  $: query = useRefetchingDashboards(instanceId, deployingName);
  $: ({ data } = $query);

  $: dashboardToRedirect = deploying
    ? getDashboardToRedirect(organization, project, data, deployingName)
    : undefined;
  $: if (dashboardToRedirect) {
    void goto(dashboardToRedirect);
  }
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer
  maxWidth={800}
  title="Project dashboards"
  showTitle={data?.length > 0}
>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
