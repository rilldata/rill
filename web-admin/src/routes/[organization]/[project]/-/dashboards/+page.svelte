<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import DashboardErrored from "@rilldata/web-admin/features/dashboards/DashboardErrored.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { deploying, deployingDashboard } = data;

  $: ({
    params: { organization, project },
  } = $page);
  $: ({ instanceId } = $runtime);

  $: query = useDashboards(instanceId);
  $: ({ data: dashboards } = $query);

  $: deployingDashboards = useDeployingDashboards(
    instanceId,
    organization,
    project,
    deploying,
    deployingDashboard,
  );
  $: ({ isPending, data: deployingDashboardsData } = $deployingDashboards);
  $: ({ redirectToDashboardUrl, dashboardsReconciling, dashboardsErrored } =
    deployingDashboardsData ?? {
      redirectToDashboardUrl: null,
      dashboardsReconciling: false,
      dashboardsErrored: false,
    });

  $: if (redirectToDashboardUrl) {
    void goto(redirectToDashboardUrl);
  }
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

{#if dashboardsErrored}
  <DashboardErrored {organization} {project} />
{:else if dashboardsReconciling || (deploying && isPending)}
  <DashboardBuilding multipleDashboards />
{:else}
  <ContentContainer
    maxWidth={800}
    title="Project dashboards"
    showTitle={dashboards?.length > 0}
  >
    <div class="flex flex-col items-center gap-y-4">
      <DashboardsTable />
    </div>
  </ContentContainer>
{/if}
