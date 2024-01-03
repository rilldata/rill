<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
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
  import { useDashboards } from "../dashboards/listing/selectors";
  import { useReports } from "../scheduled-reports/selectors";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";
  import {
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
  $: organizations = createAdminServiceListOrganizations(undefined, {
    query: {
      enabled: !!$user.data?.user,
    },
  });
  $: onOrganizationPage = isOrganizationPage($page);

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
  $: dashboards = useDashboards(instanceId);
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
        onSelectMenuOption={(organization) => goto(`/${organization}`)}
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
  </ol>
</nav>
