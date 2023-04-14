<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import DeploymentStatusChip from "../home/DeploymentStatusChip.svelte";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  $: organization = $page.params.organization;
  $: organizations = createAdminServiceListOrganizations();
  $: isOrganizationPage = $page.route.id === "/[organization]";

  $: project = $page.params.project;
  $: projects = createAdminServiceListProjectsForOrganization(organization);
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  $: dashboard = $page.params.dashboard;
  $: dashboards = useDashboardNames($runtime?.instanceId);
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
        onSelectMenuOption={(project) => goto(`/${organization}/${project}`)}
      >
        <DeploymentStatusChip {organization} {project} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if dashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={dashboard}
        isCurrentPage={isDashboardPage}
        menuOptions={$dashboards.data?.length > 1 &&
          $dashboards.data.map((dash) => ({
            key: dash,
            main: dash,
            callback: () => goto(`/${organization}/${project}/${dash}`),
          }))}
        onSelectMenuOption={(dashboard) =>
          goto(`/${organization}/${project}/${dashboard}`)}
      />
    {/if}
  </ol>
</nav>
