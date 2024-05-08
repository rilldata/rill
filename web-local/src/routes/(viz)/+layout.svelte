<script lang="ts">
  import { page } from "$app/stores";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors.js";
  import DashboardCtAs from "@rilldata/web-common/features/dashboards/workspace/DashboardCTAs.svelte";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;

  $: ({ instanceId } = data);

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: dashboardsQuery = useValidDashboards(instanceId);

  $: dashboards = $dashboardsQuery.data ?? [];

  $: dashboardOptions = dashboards.reduce((map, dimension) => {
    const label = dimension.metricsView?.state?.validSpec?.title ?? "";
    const name = dimension.meta?.name?.name ?? "";

    if (label && name)
      map.set(name.toLowerCase(), { label, section: "dashboard", depth: 0 });

    return map;
  }, new Map<string, PathOption>());

  $: pathParts = [dashboardOptions];

  $: currentPath = [dashboardName];
</script>

<div class="flex flex-col size-full">
  <header class="py-3 w-full bg-white flex gap-x-2 items-center px-4 border-b">
    {#if $dashboardsQuery.data}
      <Breadcrumbs {pathParts} {currentPath}>
        <a href="/" slot="icon">
          <Rill />
        </a>
      </Breadcrumbs>
    {/if}
    <span class="rounded-full px-2 border text-gray-800 bg-gray-50">
      PREVIEW
    </span>
    {#if route.id?.includes("dashboard")}
      <DashboardCtAs metricViewName={dashboardName} />
    {/if}
  </header>
  <slot />
</div>
