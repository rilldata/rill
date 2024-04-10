<script lang="ts">
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import Home from "@rilldata/web-common/components/icons/Home.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import ShareDashboardButton from "../dashboards/share/ShareDashboardButton.svelte";
  import { isErrorStoreEmpty } from "../errors/error-store";
  import ShareProjectButton from "../projects/ShareProjectButton.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useReports } from "../scheduled-reports/selectors";
  import OrganizationAvatar from "@rilldata/web-common/components/navigation/breadcrumbs/OrganizationAvatar.svelte";
  import { isDashboardPage, isProjectPage } from "./nav-utils";
  import Breadcrumbs, {
    type Entry,
  } from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import { useAlerts } from "../alerts/selectors";

  const user = createAdminServiceGetCurrentUser();

  $: instanceId = $runtime?.instanceId;

  $: organization = $page.params.organization || null;
  $: project = $page.params.project || null;
  $: dashboard = $page.params.dashboard || null;
  $: alert = $page.params.alert || null;
  $: report = $page.params.report || null;

  $: onProjectPage = isProjectPage($page);
  $: onDashboardPage = isDashboardPage($page);

  $: organizationQuery = createAdminServiceListOrganizations(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
      },
    },
  );

  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organizations.length,
      },
    },
  );

  $: alertsQuery = useAlerts(instanceId);
  $: reportsQuery = useReports(instanceId);

  $: organizations = $organizationQuery.data?.organizations ?? [];
  $: projects = $projectsQuery.data?.projects ?? [];
  $: alerts = $alertsQuery.data?.resources ?? [];
  $: reports = $reportsQuery.data?.resources ?? [];

  $: organizationOptions = $organizationQuery.data?.organizations.reduce(
    (map, org) => {
      map.set(org.name, { label: org.name, href: `/${org.name}` });
      return map;
    },
    new Map<string, Entry>(),
  );

  $: projectOptions = projects.reduce((map, proj) => {
    map.set(proj.name, {
      label: proj.name,
      href: `/${organization}/${proj.name}`,
    });
    return map;
  }, new Map<string, Entry>());

  $: alertOptions = alerts.reduce((map, alert) => {
    map.set(alert.meta.name.name, {
      label: alert.alert.spec.title || alert.meta.name.name,
      href: `/${organization}/${project}/-/alerts/${alert.meta.name.name}`,
    });
    return map;
  }, new Map<string, Entry>());

  $: reportOptions = reports.reduce((map, report) => {
    map.set(report.meta.name.name, {
      label: report.report.spec.title || report.meta.name.name,
      href: `/${organization}/${project}/-/reports/${report.meta.name.name}`,
    });
    return map;
  }, new Map<string, Entry>());

  $: levels = [
    organizationOptions,
    projectOptions,
    report ? reportOptions : alert ? alertOptions : null,
  ];

  $: selections = [organization, project, report || alert];
</script>

<div
  class="grid items-center w-full justify-stretch pr-4 pl-2 py-1"
  class:border-b={!onProjectPage}
  style:grid-template-columns="max-content auto max-content"
>
  <Tooltip distance={2}>
    <a
      href="/"
      class=" hover:bg-gray-200 grid place-content-center rounded p-2"
    >
      <Home color="black" size="20px" />
    </a>
    <TooltipContent slot="tooltip-content">Home</TooltipContent>
  </Tooltip>
  {#if $isErrorStoreEmpty && organization}
    <Breadcrumbs {levels} {selections}>
      <OrganizationAvatar {organization} slot="icon" />
    </Breadcrumbs>
  {:else}
    <div />
  {/if}
  <div class="flex gap-x-2 items-center">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage}
      <ShareProjectButton {organization} {project} />
    {/if}
    {#if onDashboardPage}
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
