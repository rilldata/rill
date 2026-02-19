<script lang="ts">
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { useDashboards } from "@rilldata/web-common/features/dashboards/listing/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let limit: number | undefined = undefined;
  export let showSearch = false;
  export let showSeeAll = false;
  export let seeAllHref = "/preview";

  $: dashboardsQuery = useDashboards($runtime.instanceId);
  $: ({ data: dashboardsData, isLoading, isError, error } = $dashboardsQuery);

  function getHref(name: string, isMetricsExplorer: boolean): string {
    const slug = isMetricsExplorer ? "explore" : "canvas";
    return `/${slug}/${name}`;
  }
</script>

<DashboardsTable
  data={dashboardsData ?? []}
  {isLoading}
  {isError}
  {error}
  isPreview={!!limit}
  previewLimit={limit ?? 5}
  {getHref}
  seeAllHref={showSeeAll ? seeAllHref : ""}
  toolbar={showSearch}
>
  <svelte:fragment slot="empty">
    <ResourceListEmptyState
      icon={ExploreIcon}
      message="You don't have any dashboards yet"
    >
      <span slot="action">
        <a href="/deploy" class="text-primary-600 hover:text-primary-700">
          Deploy your project
        </a>
        to share dashboards with your team
      </span>
    </ResourceListEmptyState>
  </svelte:fragment>
</DashboardsTable>
