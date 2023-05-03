<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createRuntimeServiceListCatalogEntries,
    createRuntimeServiceListFiles,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceGetProject,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { getDashboardListItemsFromFilesAndCatalogEntries } from "../projects/dashboards";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  const user = createAdminServiceGetCurrentUser();

  $: orgName = $page.params.organization;
  $: organization = createAdminServiceGetOrganization(orgName);
  $: organizations = createAdminServiceListOrganizations(undefined, {
    query: {
      enabled: !!$user.data.user,
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
        enabled: !!$organization.data.organization,
      },
    }
  );
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  // Here, we compose the dashboard list via two separate runtime queries.
  // We should create a custom hook to hide this complexity.
  $: dashboardFiles = createRuntimeServiceListFiles(
    $runtime?.instanceId,
    {
      glob: "dashboards/*.yaml",
    },
    {
      query: {
        enabled: !!projectName && !!$runtime?.instanceId,
      },
    }
  );
  $: dashboardCatalogEntries = createRuntimeServiceListCatalogEntries(
    $runtime?.instanceId,
    {
      type: "OBJECT_TYPE_METRICS_VIEW",
    },
    {
      query: {
        placeholderData: undefined,
        enabled: !!projectName && !!$runtime?.instanceId,
      },
    }
  );
  $: dashboardListItems = getDashboardListItemsFromFilesAndCatalogEntries(
    $dashboardFiles.data?.paths,
    $dashboardCatalogEntries.data?.entries
  );
  $: currentDashboard = dashboardListItems?.find(
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
        onSelectMenuOption={(project) => goto(`/${orgName}/${project}`)}
      />
    {/if}
    {#if currentDashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={currentDashboard?.title || currentDashboard.name}
        isCurrentPage={isDashboardPage}
        menuOptions={dashboardListItems?.length > 1 &&
          dashboardListItems.map((listing) => {
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
