<script lang="ts">
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import { useDashboards } from "@rilldata/web-common/features/dashboards/listing/selectors";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Build the destination URL for a dashboard row click. Defaults to the
   * relative `/{kind}/{name}` form used by the local app; cloud callers
   * provide a branch-aware builder.
   */
  export let getHref: (name: string, isMetricsExplorer: boolean) => string = (
    name,
    isMetricsExplorer,
  ) => `/${isMetricsExplorer ? "explore" : "canvas"}/${name}`;
  export let toolbar: boolean = true;
  export let emptyMessage = "No dashboards found";

  const runtimeClient = useRuntimeClient();
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: ({ data: dashboardsData, isLoading, isError, error } = $dashboardsQuery);
</script>

<ContentContainer title="Dashboards" maxWidth={1100}>
  <DashboardsTable
    data={dashboardsData ?? []}
    {isLoading}
    {isError}
    {error}
    {getHref}
    {toolbar}
  >
    <svelte:fragment slot="empty">
      <ResourceListEmptyState icon={ExploreIcon} message={emptyMessage}>
        <span slot="action">
          <slot name="empty-action" />
        </span>
      </ResourceListEmptyState>
    </svelte:fragment>
  </DashboardsTable>
</ContentContainer>
