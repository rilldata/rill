<script lang="ts">
  import { page } from "$app/stores";
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useDashboards, useIsInitialBuild } from "./selectors";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  const runtimeClient = useRuntimeClient();
  $: ({
    params: { organization, project },
  } = $page);

  $: dashboards = useDashboards(runtimeClient);
  $: ({ data: dashboardsData, isLoading, isError, error } = $dashboards);

  $: initialBuild = useIsInitialBuild(runtimeClient);
  $: isBuilding = $initialBuild.data === true;

  function getHref(name: string, isMetricsExplorer: boolean): string {
    const slug = isMetricsExplorer ? "explore" : "canvas";
    return isEmbedded
      ? `/-/embed/${slug}/${name}`
      : `/${organization}/${project}/${slug}/${name}`;
  }
</script>

<DashboardsTable
  data={dashboardsData ?? []}
  isLoading={isLoading || isBuilding}
  {isError}
  {error}
  {isPreview}
  {previewLimit}
  {getHref}
  seeAllHref={`/${organization}/${project}/-/dashboards`}
/>
