<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import Home from "@rilldata/web-common/components/icons/Home.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import OrganizationAvatar from "@rilldata/web-common/components/navigation/breadcrumbs/OrganizationAvatar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations as listOrgs,
    createAdminServiceListProjectsForOrganization as listProjects,
  } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import { useAlerts } from "../alerts/selectors";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import ShareDashboardButton from "../dashboards/share/ShareDashboardButton.svelte";
  import { isErrorStoreEmpty } from "../errors/error-store";
  import ShareProjectButton from "../projects/ShareProjectButton.svelte";
  import { useReports } from "../scheduled-reports/selectors";
  import { isMetricsExplorerPage, isProjectPage } from "./nav-utils";

  const user = createAdminServiceGetCurrentUser();

  $: instanceId = $runtime?.instanceId;

  // These can be undefined
  $: ({ organization, project, dashboard, alert, report } = $page.params);

  $: onProjectPage = isProjectPage($page);
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);

  $: organizationQuery = listOrgs(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
      },
    },
  );

  $: projectsQuery = listProjects(organization, undefined, {
    query: {
      enabled: !!organizations.length,
    },
  });

  $: dashboardsQuery = useValidDashboards(instanceId);

  $: alertsQuery = useAlerts(instanceId);
  $: reportsQuery = useReports(instanceId);

  $: organizations = $organizationQuery.data?.organizations ?? [];
  $: projects = $projectsQuery.data?.projects ?? [];
  $: dashboards = $dashboardsQuery.data ?? [];
  $: alerts = $alertsQuery.data?.resources ?? [];
  $: reports = $reportsQuery.data?.resources ?? [];

  $: organizationPaths = organizations.reduce(
    (map, { name }) => map.set(name, { label: name }),
    new Map<string, PathOption>(),
  );

  $: projectPaths = projects.reduce((map, { name: label }) => {
    return map.set(label, { label });
  }, new Map<string, PathOption>());

  $: dashboardPaths = dashboards.reduce((map, { meta, metricsView }) => {
    const id = meta.name.name;
    return map.set(id, {
      label: metricsView?.state?.validSpec?.title || id,
    });
  }, new Map<string, PathOption>());

  $: alertPaths = alerts.reduce((map, alert) => {
    const id = alert.meta.name.name;
    return map.set(id, {
      label: alert.alert.spec.title || id,
      section: "-/alerts",
    });
  }, new Map<string, PathOption>());

  $: reportPaths = reports.reduce((map, report) => {
    const id = report.meta.name.name;
    return map.set(id, {
      label: report.report.spec.title || id,
      section: "-/reports",
    });
  }, new Map<string, PathOption>());

  $: pathParts = [
    organizationPaths,
    projectPaths,
    dashboardPaths,
    report ? reportPaths : alert ? alertPaths : null,
  ];

  $: currentPath = [organization, project, dashboard, report || alert];
</script>

<div
  class="flex items-center w-full pr-4 pl-2 py-1"
  class:border-b={!onProjectPage}
>
  <Tooltip distance={2}>
    <a href="/" class="hover:bg-gray-200 grid place-content-center rounded p-2">
      <Home color="black" size="20px" />
    </a>
    <TooltipContent slot="tooltip-content">Home</TooltipContent>
  </Tooltip>
  {#if $isErrorStoreEmpty && organization}
    <Breadcrumbs {pathParts} {currentPath}>
      <OrganizationAvatar {organization} slot="icon" />
    </Breadcrumbs>
  {/if}
  <div class="flex gap-x-2 items-center ml-auto">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage}
      <ShareProjectButton {organization} {project} />
    {/if}
    {#if onMetricsExplorerPage}
      <StateManagersProvider metricsViewName={dashboard}>
        <LastRefreshedDate {dashboard} />
        {#if $user.isSuccess && $user.data.user}
          <CreateAlert />
          <Bookmarks />
        {/if}
        <ShareDashboardButton />
      </StateManagersProvider>
    {/if}
    {#if $user.isSuccess}
      {#if $user.data && $user.data.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</div>
