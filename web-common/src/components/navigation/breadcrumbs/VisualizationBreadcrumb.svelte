<script lang="ts">
  import { page } from "$app/stores";
  import { useValidVisualizations } from "@rilldata/web-common/features/dashboards/selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type {
    V1MetricsViewSpec,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { isAnyDashboardPage, isMetricsExplorerPage } from "../nav-utils";
  import BreadcrumbItem, {
    type BreadcrumbMenuItem,
  } from "./BreadcrumbItem.svelte";

  let currentResource: V1Resource;
  let currentMetricsExplorer: V1MetricsViewSpec;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;
  $: instanceId = $runtime?.instanceId;

  // All visualizations
  $: visualizations = useValidVisualizations(instanceId);

  // Current visualization
  $: currentResource = $visualizations?.data?.find(
    (resource) => resource.meta.name.name === $page.params.dashboard,
  );
  $: currentVizName = currentResource?.meta?.name?.name;
  $: currentMetricsExplorer = currentResource?.metricsView?.state?.validSpec;
  $: currentCustomDashboard = currentResource?.dashboard?.spec;

  // Navigation helpers
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);
  $: onDashboardPage = isAnyDashboardPage($page);

  $: menuItems =
    $visualizations?.data?.length > 1 &&
    $visualizations.data.map((resource) => {
      const isMetricsExplorer = !!resource?.metricsView;
      return {
        key: resource.meta.name.name,
        main: isMetricsExplorer
          ? resource?.metricsView?.state?.validSpec?.title ||
            resource.meta.name.name
          : resource?.dashboard?.spec?.title || resource.meta.name.name,
        kind: isMetricsExplorer
          ? ResourceKind.MetricsView
          : ResourceKind.Dashboard,
      };
    });

  function makeMenuItemHref(item: BreadcrumbMenuItem) {
    return item.kind === ResourceKind.MetricsView
      ? `/${orgName}/${projectName}/${item.key}`
      : `/${orgName}/${projectName}/-/dashboards/${item.key}`;
  }
</script>

{#if currentResource}
  <span class="text-gray-600">/</span>
  <BreadcrumbItem
    label={onMetricsExplorerPage
      ? currentMetricsExplorer?.title || currentVizName
      : currentCustomDashboard?.title || currentVizName}
    href={onMetricsExplorerPage
      ? `/${orgName}/${projectName}/${currentVizName}`
      : `/${orgName}/${projectName}/-/dashboards/${currentVizName}`}
    {menuItems}
    menuKey={currentVizName}
    {makeMenuItemHref}
    isCurrentPage={onDashboardPage}
  />
{/if}
