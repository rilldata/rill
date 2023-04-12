<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceListOrganizations,
    createAdminServiceListProjects,
  } from "../../client";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  $: organization = $page.params.organization;
  $: organizations = createAdminServiceListOrganizations();
  $: organizationPageActive = $page.route.id === "/[organization]";

  $: project = $page.params.project;
  $: projects = createAdminServiceListProjects(organization);
  $: projectPageActive = $page.route.id === "/[organization]/[project]";

  $: dashboard = $page.params.dashboard;
  $: dashboards = useDashboardNames($runtime?.instanceId);
  $: dashboardPageActive =
    $page.route.id === "/[organization]/[project]/[dashboard]";
</script>

<nav>
  <ol class="flex flex-row items-center">
    {#if organization}
      <BreadcrumbItem
        label={organization}
        isActive={organizationPageActive}
        options={$organizations.data?.organizations?.length > 1 &&
          $organizations.data.organizations.map((org) => ({
            main: org.name,
            callback: () => goto(`/${org.name}`),
          }))}
      >
        <OrganizationAvatar {organization} slot="icon" />
      </BreadcrumbItem>
    {/if}
    {#if project}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={project}
        isActive={projectPageActive}
        options={$projects.data?.projects?.length > 1 &&
          $projects.data.projects.map((proj) => ({
            main: proj.name,
            callback: () => goto(`/${organization}/${proj.name}`),
          }))}
      />
    {/if}
    {#if dashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={dashboard}
        isActive={dashboardPageActive}
        options={$dashboards.data?.length > 1 &&
          $dashboards.data.map((dash) => ({
            main: dash,
            callback: () => goto(`/${organization}/${project}/${dash}`),
          }))}
      />
    {/if}
  </ol>
</nav>
