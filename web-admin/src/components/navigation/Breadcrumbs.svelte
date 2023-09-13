<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useProjectDeploymentStatus } from "@rilldata/web-admin/components/projects/selectors";
  import { Tag } from "@rilldata/web-common/components/tag";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceGetProject,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { useDashboards } from "../projects/dashboards";
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
        enabled: !!$organization.data?.organization,
      },
    }
  );
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  $: dashboards = useDashboards(instanceId);
  $: currentDashboard = $dashboards?.data?.find(
    (listing) => listing.name === $page.params.dashboard
  );
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
        isCurrentPage={isProjectPage}
      />
    {/if}
    {#if currentDashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={currentDashboard?.label || currentDashboard.name}
        href={`/${orgName}/${projectName}/${currentDashboard.name}`}
        menuOptions={$dashboards?.data?.length > 1 &&
          $dashboards.data.map((listing) => {
            return {
              key: listing.name,
              main: listing?.label || listing.name,
            };
          })}
        menuKey={currentDashboard.name}
        onSelectMenuOption={(dashboard) =>
          goto(`/${orgName}/${projectName}/${dashboard}`)}
        isCurrentPage={isDashboardPage}
      />
    {/if}
    <!-- This is a temporary solution until we move intra-project navigation to tabs -->
    {#if $page.route.id.endsWith("/-/logs")}
      <Tag>Logs</Tag>
    {/if}
  </ol>
</nav>
