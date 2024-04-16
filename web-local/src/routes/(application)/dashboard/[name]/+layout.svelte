<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors.js";
  import { page } from "$app/stores";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import DashboardCtAs from "@rilldata/web-common/features/dashboards/workspace/DashboardCTAs.svelte";

  export let data;

  $: ({ instanceId } = data);

  $: dashboardName = $page.params.name;

  $: projectTitleQuery = useProjectTitle(instanceId);
  $: dashboardsQuery = useValidDashboards(instanceId);

  $: projectName = ($projectTitleQuery.data as string | undefined) ?? null;
  $: dashboards = $dashboardsQuery.data ?? [];

  $: dashboardOptions = dashboards.reduce((map, dimension) => {
    const label = dimension.metricsView?.state?.validSpec?.title ?? "";
    const id = dimension.meta?.name?.name ?? "";

    if (label && id) map.set(id, { label, section: "dashboard", depth: 0 });

    return map;
  }, new Map<string, PathOption>());

  $: projectOptions = new Map<string, PathOption>([
    [projectName ?? "", { label: projectName ?? "", href: "/" }],
  ]);

  $: pathParts = [projectOptions, dashboardOptions];

  $: currentPath = [projectName ?? "", dashboardName];
</script>

<div class="flex flex-col size-full">
  <header class="py-3 w-full bg-white flex gap-x-2 items-center px-4 border-b">
    {#if $dashboardsQuery.data}
      <Breadcrumbs {pathParts} {currentPath}>
        <Rill slot="icon" />
      </Breadcrumbs>
    {/if}
    <span class="rounded-full px-2 border text-gray-800 bg-gray-50">
      PREVIEW
    </span>
    <DashboardCtAs metricViewName={dashboardName} />
  </header>
  <slot />
</div>
