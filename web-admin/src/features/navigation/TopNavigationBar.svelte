<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import ShareDashboardButton from "@rilldata/web-admin/features/dashboards/share/ShareDashboardButton.svelte";
  import UserInviteButton from "@rilldata/web-admin/features/projects/user-invite/UserInviteButton.svelte";
  import { useShareableURLMetricsView } from "@rilldata/web-admin/features/shareable-urls/selectors";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useValidVisualizations } from "@rilldata/web-common/features/dashboards/selectors";
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
  import { useReports } from "../scheduled-reports/selectors";
  import PageTitle from "../shareable-urls/PageTitle.svelte";
  import {
    isMagicLinkPage,
    isMetricsExplorerPage,
    isProjectPage,
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
  } = $page.params);

  $: onProjectPage = isProjectPage($page);
  $: onAlertPage = !!alert;
  $: onReportPage = !!report;
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);
  $: onMagicLinkPage = isMagicLinkPage($page);

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

  $: visualizationsQuery = useValidVisualizations(instanceId);

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
    (map, { name }) => map.set(name.toLowerCase(), { label: name }),
    new Map<string, PathOption>(),
  );

  $: projectPaths = projects.reduce(
    (map, { name }) => map.set(name.toLowerCase(), { label: name }),
    new Map<string, PathOption>(),
  );

  $: visualizationPaths = visualizations.reduce(
    (map, { meta, metricsView, dashboard }) => {
      const name = meta.name.name;
      const isMetricsExplorer = !!metricsView;
      return map.set(name.toLowerCase(), {
        label:
          (isMetricsExplorer
            ? metricsView?.state?.validSpec?.title
            : dashboard?.spec?.title) || name,
        section: isMetricsExplorer ? undefined : "-/dashboards",
      });
    },
    new Map<string, PathOption>(),
  );

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

  // When visiting a magic link, the metrics view name won't be in the URL. However, the URL's token will
  // have access to only one metrics view. So, we can get the metrics view name from the first (and only) metrics view resource.
  $: metricsViewQuery = useShareableURLMetricsView(instanceId, onMagicLinkPage);
  $: dashboard = onMagicLinkPage
    ? $metricsViewQuery.data?.meta?.name?.name
    : dashboardParam;

  $: magicLinkDashboardTitle =
    $metricsViewQuery.data?.metricsView?.spec?.title ?? dashboard ?? "";

  $: currentPath = [organization, project, dashboard, report || alert];
</script>

<div
  class="flex items-center w-full pr-4 pl-2 py-1"
  class:border-b={!onProjectPage}
>
  <!-- Left side -->
  <a
    href={rillLogoHref}
    class="hover:bg-gray-200 grid place-content-center rounded p-2"
  >
    <Rill />
  </a>
  {#if onMagicLinkPage}
    <PageTitle title={magicLinkDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <!-- Right side -->
  <div class="flex gap-x-2 items-center ml-auto">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage && manageProjectMembers}
      <UserInviteButton />
    {/if}
    {#if onMetricsExplorerPage || onMagicLinkPage}
      <StateManagersProvider metricsViewName={dashboard}>
        <LastRefreshedDate {dashboard} />
        <GlobalDimensionSearch metricsViewName={dashboard} />
        {#if $user.isSuccess && $user.data.user && !onMagicLinkPage}
          <CreateAlert />
          <Bookmarks metricsViewName={dashboard} />
          <ShareDashboardButton {createMagicAuthTokens} />
        {/if}
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
