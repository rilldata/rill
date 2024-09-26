<script lang="ts">
  import { page } from "$app/stores";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { getBreadcrumbOptions } from "@rilldata/web-common/features/dashboards/dashboard-utils";
  import {
    useValidCanvases,
    useValidExplores,
  } from "@rilldata/web-common/features/dashboards/selectors.js";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardCtAs from "@rilldata/web-common/features/dashboards/workspace/DashboardCTAs.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: exploresQuery = useValidExplores(instanceId);
  $: canvasQuery = useValidCanvases(instanceId);
  $: projectTitleQuery = useProjectTitle(instanceId);

  $: projectTitle = $projectTitleQuery.data ?? "Untitled Rill Project";

  $: explores = $exploresQuery.data ?? [];
  $: canvases = $canvasQuery.data ?? [];

  $: dashboardOptions = getBreadcrumbOptions(explores, canvases);

  $: projectPath = <PathOption>{
    label: projectTitle,
    section: "project",
    depth: -1,
    href: "/",
  };

  $: pathParts = [
    new Map([[projectTitle.toLowerCase(), projectPath]]),
    dashboardOptions,
  ];

  $: currentPath = [projectTitle, dashboardName.toLowerCase()];

  $: currentDashboard = explores.find(
    (d) => d.meta?.name?.name?.toLowerCase() === dashboardName.toLowerCase(),
  );

  $: metricsViewName = currentDashboard?.meta?.name?.name;
</script>

<div class="flex flex-col size-full overflow-hidden">
  <header>
    {#if $exploresQuery.data}
      <Breadcrumbs {pathParts} {currentPath}>
        <a href="/" slot="icon">
          <Rill />
        </a>
      </Breadcrumbs>
    {/if}
    <span class="rounded-full px-2 border text-gray-800 bg-gray-50">
      PREVIEW
    </span>
    {#if route.id?.includes("explore") && metricsViewName}
      <StateManagersProvider {metricsViewName} exploreName={dashboardName}>
        <DashboardCtAs exploreName={dashboardName} />
      </StateManagersProvider>
    {/if}
  </header>
  <slot />
</div>

<style lang="postcss">
  header {
    @apply w-full bg-background box-border;
    @apply flex gap-x-2 items-center px-4 border-b flex-none;
    height: var(--header-height);
  }
</style>
