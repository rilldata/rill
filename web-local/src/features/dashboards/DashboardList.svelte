<script lang="ts">
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { useDashboards } from "@rilldata/web-common/features/dashboards/listing/selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";

  export let limit: number | undefined = undefined;
  export let showSearch = false;
  export let showSeeAll = false;
  export let seeAllHref = "/dashboards";

  const runtimeClient = useRuntimeClient();
  $: dashboardsQuery = useDashboards(runtimeClient);
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
    {#if $previewModeStore}
      <ResourceListEmptyState icon={ExploreIcon} message="No dashboards found">
        <span slot="action">
          Create dashboards using your code editor, then return here to preview
          them.
        </span>
      </ResourceListEmptyState>
    {:else}
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
    {/if}
  </svelte:fragment>
</DashboardsTable>
