<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import type {
    V1MetricsViewSpec,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { isAnyDashboardPage } from "../nav-utils";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  $: instanceId = $runtime?.instanceId;

  $: dashboards = useValidDashboards(instanceId);
  let currentResource: V1Resource;
  $: currentResource = $dashboards?.data?.find(
    (listing) => listing.meta.name.name === $page.params.dashboard,
  );
  $: currentDashboardName = currentResource?.meta?.name?.name;
  let currentDashboard: V1MetricsViewSpec;
  $: currentDashboard = currentResource?.metricsView?.state?.validSpec;
  $: onDashboardPage = isAnyDashboardPage($page);
</script>

{#if currentDashboard}
  <span class="text-gray-600">/</span>
  <BreadcrumbItem
    label={currentDashboard?.title || currentDashboardName}
    href={`/${orgName}/${projectName}/${currentDashboardName}`}
    menuOptions={$dashboards?.data?.length > 1 &&
      $dashboards.data.map((listing) => {
        return {
          key: listing.meta.name.name,
          main:
            listing?.metricsView?.state?.validSpec?.title ||
            listing.meta.name.name,
        };
      })}
    menuKey={currentDashboardName}
    onSelectMenuOption={(dashboard) =>
      goto(`/${orgName}/${projectName}/${dashboard}`)}
    isCurrentPage={onDashboardPage}
  />
{/if}
