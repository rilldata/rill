<script lang="ts">
  import { page } from "$app/stores";
  import Banner from "@rilldata/web-admin/features/billing/banner/Banner.svelte";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import ShareDashboardButton from "@rilldata/web-admin/features/dashboards/share/ShareDashboardButton.svelte";
  import UserInviteButton from "@rilldata/web-admin/features/projects/user-invite/UserInviteButton.svelte";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
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
  import { useDashboardsV2 } from "../dashboards/listing/selectors";
  import PageTitle from "../public-urls/PageTitle.svelte";
  import { createAdminServiceGetMagicAuthToken } from "../public-urls/get-magic-auth-token";
  import { usePublicURLMetricsView } from "../public-urls/selectors";
  import { useReports } from "../scheduled-reports/selectors";
  import {
    isMetricsExplorerPage,
    isOrganizationPage,
    isProjectPage,
    isPublicURLPage,
  } from "./nav-utils";

  export let createMagicAuthTokens: boolean;
  export let manageProjectMembers: boolean;

  const user = createAdminServiceGetCurrentUser();

  $: instanceId = $runtime?.instanceId;

  // These can be undefined
  $: ({
    organization,
    project,
    dashboard: dashboardParam,
    alert,
    report,
    token,
  } = $page.params);

  $: onProjectPage = isProjectPage($page);
  $: onAlertPage = !!alert;
  $: onReportPage = !!report;
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);
  $: onPublicURLPage = isPublicURLPage($page);
  $: onOrgPage = isOrganizationPage($page);

  $: loggedIn = !!$user.data?.user;
  $: rillLogoHref = !loggedIn ? "https://www.rilldata.com" : "/";

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
      enabled: !!organization,
    },
  });

  $: visualizationsQuery = useDashboardsV2(instanceId);

  $: alertsQuery = useAlerts(instanceId, onAlertPage);
  $: reportsQuery = useReports(instanceId, onReportPage);

  $: organizations =
    $organizationQuery.data?.organizations ??
    // handle case when visiting root cloud page directly (ui.rilldata.com)
    (organization ? [{ name: organization, id: organization }] : []);
  $: projects = $projectsQuery.data?.projects ?? [];
  $: visualizations = $visualizationsQuery.data ?? [];
  $: alerts = $alertsQuery.data?.resources ?? [];
  $: reports = $reportsQuery.data?.resources ?? [];

  $: organizationPaths = organizations.reduce(
    (map, { name, displayName }) =>
      map.set(name.toLowerCase(), { label: displayName || name }),
    new Map<string, PathOption>(),
  );

  $: projectPaths = projects.reduce(
    (map, { name }) => map.set(name.toLowerCase(), { label: name }),
    new Map<string, PathOption>(),
  );

  $: visualizationPaths = visualizations.reduce((map, { resource }) => {
    const name = resource.meta.name.name;
    const isMetricsExplorer = !!resource?.metricsView;
    return map.set(name.toLowerCase(), {
      label:
        (isMetricsExplorer
          ? resource?.metricsView?.spec?.title
          : resource?.dashboard?.spec?.title) || name,
      section: isMetricsExplorer ? undefined : "-/dashboards",
    });
  }, new Map<string, PathOption>());

  $: alertPaths = alerts.reduce((map, alert) => {
    const name = alert.meta.name.name;
    return map.set(name.toLowerCase(), {
      label: alert.alert.spec.title || name,
      section: "-/alerts",
    });
  }, new Map<string, PathOption>());

  $: reportPaths = reports.reduce((map, report) => {
    const name = report.meta.name.name;
    return map.set(name.toLowerCase(), {
      label: report.report.spec.title || name,
      section: "-/reports",
    });
  }, new Map<string, PathOption>());

  $: pathParts = [
    organizationPaths,
    projectPaths,
    visualizationPaths,
    report ? reportPaths : alert ? alertPaths : null,
  ];

  $: dashboardQuery = useDashboard(instanceId, dashboardParam, {
    enabled: !!instanceId && onMetricsExplorerPage,
  });
  $: isDashboardValid = !!$dashboardQuery.data?.metricsView?.state?.validSpec;

  // Public URLs do not have the metrics view name in the URL. However, the magic token's metadata includes the metrics view name.
  $: tokenQuery = createAdminServiceGetMagicAuthToken(token);
  $: dashboard = onPublicURLPage
    ? $tokenQuery?.data?.token?.metricsView
    : dashboardParam;

  // If on a Public URL, get the dashboard title
  $: metricsViewQuery = usePublicURLMetricsView(
    instanceId,
    $tokenQuery?.data?.token?.metricsView,
    onPublicURLPage,
  );
  $: publicURLDashboardTitle =
    $metricsViewQuery.data?.metricsView?.spec?.title ?? dashboard ?? "";

  $: currentPath = [organization, project, dashboard, report || alert];
</script>

{#if organization}
  <Banner {organization} />
{/if}
<div
  class="flex items-center w-full pr-4 pl-2 py-1"
  class:border-b={!onProjectPage && !onOrgPage}
>
  <!-- Left side -->
  <a
    href={rillLogoHref}
    class="hover:bg-gray-200 grid place-content-center rounded p-2"
  >
    <Rill />
  </a>
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <!-- Right side -->
  <div class="flex gap-x-2 items-center ml-auto">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage && manageProjectMembers}
      <UserInviteButton {organization} {project} />
    {/if}
    {#if (onMetricsExplorerPage && isDashboardValid) || onPublicURLPage}
      {#key dashboard}
        <StateManagersProvider metricsViewName={dashboard}>
          <LastRefreshedDate {dashboard} />
          <GlobalDimensionSearch metricsViewName={dashboard} />
          {#if $user.isSuccess && $user.data.user && !onPublicURLPage}
            <CreateAlert />
            <Bookmarks metricsViewName={dashboard} />
            <ShareDashboardButton {createMagicAuthTokens} />
          {/if}
        </StateManagersProvider>
      {/key}
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
