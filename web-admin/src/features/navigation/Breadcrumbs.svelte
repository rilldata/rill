<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import type {
    V1MetricsViewSpec,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceGetProject,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { getActiveOrgLocalStorageKey } from "../organizations/active-org/local-storage";
  import { useReports } from "../scheduled-reports/selectors";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";
  import {
    isAlertPage,
    isDashboardPage,
    isOrganizationPage,
    isProjectPage,
    isReportPage,
  } from "./nav-utils";

  const user = createAdminServiceGetCurrentUser();

  $: instanceId = $runtime?.instanceId;

  // Org breadcrumb
  $: orgName = $page.params.organization;
  $: organization = createAdminServiceGetOrganization(orgName);
  $: organizations = createAdminServiceListOrganizations(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
      },
    },
  );
  $: onOrganizationPage = isOrganizationPage($page);
  async function onOrgChange(org: string) {
    const activeOrgLocalStorageKey = getActiveOrgLocalStorageKey(
      $user.data?.user?.id,
    );
    localStorage.setItem(activeOrgLocalStorageKey, org);
    await goto(`/${org}`);
  }

  // Project breadcrumb
  $: projectName = $page.params.project;
  $: project = createAdminServiceGetProject(orgName, projectName);
  $: projects = createAdminServiceListProjectsForOrganization(
    orgName,
    undefined,
    {
      query: {
        enabled: !!$organization.data?.organization,
      },
    },
  );
  $: onProjectPage = isProjectPage($page);

  // Dashboard breadcrumb
  $: dashboards = useValidDashboards(instanceId);
  let currentResource: V1Resource;
  $: currentResource = $dashboards?.data?.find(
    (listing) => listing.meta.name.name === $page.params.dashboard,
  );
  $: currentDashboardName = currentResource?.meta?.name?.name;
  let currentDashboard: V1MetricsViewSpec;
  $: currentDashboard = currentResource?.metricsView?.state?.validSpec;
  $: onDashboardPage = isDashboardPage($page);

  // Report breadcrumb
  $: reportName = $page.params.report;
  $: reports = useReports(instanceId);
  $: onReportPage = isReportPage($page);

  // Alert breadcrumb
  $: alertName = $page.params.alert;
  $: alerts = useAlerts(instanceId);
  $: onAlertPage = isAlertPage($page);
</script>

<nav>
  <ol class="flex flex-row items-center">
    {#if $organization.data?.organization}
      <BreadcrumbItem
        label={orgName}
        href={`/${orgName}`}
        menuOptions={$organizations.data?.organizations?.length > 1 &&
          $organizations.data.organizations.map((org) => ({
            key: org.name,
            main: org.name,
          }))}
        menuKey={orgName}
        onSelectMenuOption={onOrgChange}
        isCurrentPage={onOrganizationPage}
      >
        <OrganizationAvatar organization={orgName} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if $project.data?.project}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={projectName}
        href={onReportPage
          ? `/${orgName}/${projectName}/-/reports`
          : `/${orgName}/${projectName}`}
        menuOptions={$projects.data?.projects?.length > 1 &&
          $projects.data.projects.map((proj) => ({
            key: proj.name,
            main: proj.name,
          }))}
        menuKey={projectName}
        onSelectMenuOption={(project) => goto(`/${orgName}/${project}`)}
        isCurrentPage={onProjectPage}
      />
    {/if}
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
    {#if reportName}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={reportName}
        href={`/${orgName}/${projectName}/-/reports/${reportName}`}
        menuOptions={$reports.data?.resources.map((resource) => ({
          key: resource.meta.name.name,
          main: resource.report.spec.title || resource.meta.name.name,
        }))}
        menuKey={reportName}
        onSelectMenuOption={(report) =>
          goto(`/${orgName}/${projectName}/-/reports/${report}`)}
        isCurrentPage={onReportPage}
      />
    {/if}
    {#if alertName}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={alertName}
        href={`/${orgName}/${projectName}/-/alerts/${alertName}`}
        menuOptions={$alerts.data?.resources.map((resource) => ({
          key: resource.meta.name.name,
          main: resource.alert.spec.title || resource.meta.name.name,
        }))}
        menuKey={alertName}
        onSelectMenuOption={(alert) =>
          goto(`/${orgName}/${projectName}/-/alerts/${alert}`)}
        isCurrentPage={onAlertPage}
      />
    {/if}
  </ol>
</nav>
