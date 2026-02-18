<script lang="ts">
  import { page } from "$app/stores";
  import SharedDashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useDashboards } from "./selectors";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  $: ({ instanceId } = $runtime);
  $: ({
    params: { organization, project },
  } = $page);

  $: dashboards = useDashboards(instanceId);
  $: ({
    data: dashboardsData,
    isLoading,
    isError,
    error,
  } = $dashboards);

  function getHref(name: string, isMetricsExplorer: boolean): string {
    const slug = isMetricsExplorer ? "explore" : "canvas";
    return isEmbedded
      ? `/-/embed/${slug}/${name}`
      : `/${organization}/${project}/${slug}/${name}`;
  }

  $: seeAllHref = `/${organization}/${project}/-/dashboards`;
</script>

<SharedDashboardsTable
  data={dashboardsData ?? []}
  {isLoading}
  {isError}
  {error}
  {isPreview}
  {previewLimit}
  {getHref}
  {seeAllHref}
/>
