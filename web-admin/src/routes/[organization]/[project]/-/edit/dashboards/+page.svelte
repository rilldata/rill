<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import { useDashboards } from "@rilldata/web-common/features/dashboards/listing/selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: ({ data: dashboardsData, isLoading, isError, error } = $dashboardsQuery);
</script>

<ContentContainer title="Dashboards" maxWidth={1100}>
  <DashboardsTable data={dashboardsData ?? []} {isLoading} {isError} {error} />
</ContentContainer>
