<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useProjectDeploymentStatus } from "@rilldata/web-admin/components/projects/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceGetProject,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { useDashboardListItems } from "../projects/dashboards";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
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
  // Poll specifically for the project's deployment status
  $: projectDeploymentStatus = useProjectDeploymentStatus(orgName, projectName);
  $: projects = createAdminServiceListProjectsForOrganization(
    orgName,
    undefined,
    {
      query: {
        enabled: !!$organization.data.organization,
      },
    }
  );
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  $: dashboardListItems = useDashboardListItems(
    instanceId,
    $projectDeploymentStatus.data
  );
  $: currentDashboard = $dashboardListItems?.items?.find(
    (listing) => listing.name === $page.params.dashboard
  );
  $: isDashboardPage =
    $page.route.id === "/[organization]/[project]/[dashboard]";
</script>

<nav>
  <ol class="flex flex-row items-center">
    {#if $organization.data.organization}
      <BreadcrumbItem
        label={orgName}
        isCurrentPage={isOrganizationPage}
        menuOptions={$organizations.data?.organizations?.length > 1 &&
          $organizations.data.organizations.map((org) => ({
            key: org.name,
            main: org.name,
          }))}
        menuKey={orgName}
        onSelectMenuOption={(organization) => goto(`/${organization}`)}
      >
        <OrganizationAvatar organization={orgName} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if $project.data.project}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={projectName}
        isCurrentPage={isProjectPage}
        menuOptions={$projects.data?.projects?.length > 1 &&
          $projects.data.projects.map((proj) => ({
            key: proj.name,
            main: proj.name,
          }))}
        menuKey={projectName}
        onSelectMenuOption={(project) =>
          goto(`/${orgName}/${project}/-/redirect`)}
      />
    {/if}
    {#if currentDashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={currentDashboard?.title || currentDashboard.name}
        isCurrentPage={isDashboardPage}
        menuOptions={$dashboardListItems?.items?.length > 1 &&
          $dashboardListItems.items.map((listing) => {
            return {
              key: listing.name,
              main: listing?.title || listing.name,
            };
          })}
        menuKey={currentDashboard.name}
        onSelectMenuOption={(dashboard) =>
          goto(`/${orgName}/${projectName}/${dashboard}`)}
      />
    {/if}
  </ol>
</nav>
