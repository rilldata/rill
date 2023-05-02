<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useProject } from "@rilldata/web-admin/components/projects/use-project";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { useDashboardListItems } from "../projects/dashboards";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  const user = createAdminServiceGetCurrentUser();

  $: organization = $page.params.organization;
  $: organizations = createAdminServiceListOrganizations(undefined, {
    query: {
      enabled: !!$user.data.user,
    },
  });
  $: isOrganizationPage = $page.route.id === "/[organization]";

  $: project = $page.params.project;
  $: proj = useProject(organization, project);
  $: projects = createAdminServiceListProjectsForOrganization(organization);
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  // Here, we compose the dashboard list via two separate runtime queries.
  // We should create a custom hook to hide this complexity.
  $: dashboardListItems = useDashboardListItems($runtime?.instanceId, proj);
  $: currentDashboard = $dashboardListItems?.find(
    (listing) => listing.name === $page.params.dashboard
  );
  $: isDashboardPage =
    $page.route.id === "/[organization]/[project]/[dashboard]";
</script>

<nav>
  <ol class="flex flex-row items-center">
    {#if organization}
      <BreadcrumbItem
        label={organization}
        isCurrentPage={isOrganizationPage}
        menuOptions={$organizations.data?.organizations?.length > 1 &&
          $organizations.data.organizations.map((org) => ({
            key: org.name,
            main: org.name,
            callback: () => goto(`/${org.name}`),
          }))}
        menuKey={organization}
        onSelectMenuOption={(organization) => goto(`/${organization}`)}
      >
        <OrganizationAvatar {organization} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if project}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={project}
        isCurrentPage={isProjectPage}
        menuOptions={$projects.data?.projects?.length > 1 &&
          $projects.data.projects.map((proj) => ({
            key: proj.name,
            main: proj.name,
            callback: () => goto(`/${organization}/${proj.name}`),
          }))}
        menuKey={project}
        onSelectMenuOption={(project) => goto(`/${organization}/${project}`)}
      />
    {/if}
    {#if currentDashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={currentDashboard?.title || currentDashboard.name}
        isCurrentPage={isDashboardPage}
        menuOptions={$dashboardListItems?.length > 1 &&
          $dashboardListItems.map((listing) => {
            return {
              key: listing.name,
              main: listing?.title || listing.name,
              callback: () =>
                goto(`/${organization}/${project}/${listing.name}`),
            };
          })}
        menuKey={currentDashboard.name}
        onSelectMenuOption={(dashboard) =>
          goto(`/${organization}/${project}/${dashboard}`)}
      />
    {/if}
  </ol>
</nav>
