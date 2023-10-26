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
  import { useDashboards } from "../dashboards/listing/dashboards";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import { isProjectPage } from "./nav-utils";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  const user = createAdminServiceGetCurrentUser();

  $: instanceId = $runtime?.instanceId;

  $: orgName = $page.params.organization;
  $: organization = createAdminServiceGetOrganization(orgName);
  $: organizations = createAdminServiceListOrganizations(undefined, {
    query: {
      enabled: !!$user.data?.user,
    },
  });
  $: isOrganizationPage = $page.route.id === "/[organization]";

  $: projectName = $page.params.project;
  $: project = createAdminServiceGetProject(orgName, projectName);
  $: projects = createAdminServiceListProjectsForOrganization(
    orgName,
    undefined,
    {
      query: {
        enabled: !!$organization.data?.organization,
      },
    }
  );
  $: onProjectPage = isProjectPage($page);

  $: dashboards = useDashboards(instanceId);
  let currentResource: V1Resource;
  $: currentResource = $dashboards?.data?.find(
    (listing) => listing.meta.name.name === $page.params.dashboard
  );
  $: currentDashboardName = currentResource?.meta?.name?.name;
  let currentDashboard: V1MetricsViewSpec;
  $: currentDashboard = currentResource?.metricsView?.state?.validSpec;
  $: isDashboardPage =
    $page.route.id === "/[organization]/[project]/[dashboard]";
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
        isCurrentPage={isOrganizationPage}
      >
        <OrganizationAvatar organization={orgName} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if $project.data?.project}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={projectName}
        href={`/${orgName}/${projectName}`}
        menuOptions={$projects.data?.projects?.length > 1 &&
          $projects.data.projects.map((proj) => ({
            key: proj.name,
            main: proj.name,
          }))}
        menuKey={projectName}
        onSelectMenuOption={(project) =>
          goto(
            isDashboardPage
              ? `/${orgName}/${project}/-/redirect`
              : `/${orgName}/${project}`
          )}
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
        isCurrentPage={isDashboardPage}
      />
    {/if}
  </ol>
</nav>
