<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-common/features/dashboards/listing/DashboardsTable.svelte";
  import { useDashboards } from "@rilldata/web-common/features/dashboards/listing/selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: ({ data: dashboardsData, isLoading, isError, error } = $dashboardsQuery);

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: branchPart = branchPathPrefix(branch);

  function getHref(name: string, isMetricsExplorer: boolean): string {
    const slug = isMetricsExplorer ? "explore" : "canvas";
    return `/${organization}/${project}${branchPart}/${slug}/${name}`;
  }
</script>

<ContentContainer title="Dashboards" maxWidth={1100}>
  <DashboardsTable
    data={dashboardsData ?? []}
    {isLoading}
    {isError}
    {error}
    {getHref}
    toolbar
  />
</ContentContainer>
