<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createRuntimeServiceListCatalogEntries } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  $: organization = $page.params.organization;
  $: organizations = createAdminServiceListOrganizations();
  $: isOrganizationPage = $page.route.id === "/[organization]";

  $: project = $page.params.project;
  $: projects = createAdminServiceListProjectsForOrganization(organization);
  $: isProjectPage = $page.route.id === "/[organization]/[project]";

  $: dashboard = $page.params.dashboard;
  $: dashboards = createRuntimeServiceListCatalogEntries(
    $runtime?.instanceId,
    {
      type: "OBJECT_TYPE_METRICS_VIEW",
    },
    { query: { enabled: !!project && !!$runtime?.instanceId } }
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
      />
    {/if}
    {#if dashboard}
      <span class="text-gray-600">/</span>
      <BreadcrumbItem
        label={dashboard}
        isCurrentPage={isDashboardPage}
        menuOptions={$dashboards.data?.entries?.length > 1 &&
          $dashboards.data.entries.map((dash) => ({
            key: dash.name,
            main: dash.name,
            callback: () => goto(`/${organization}/${project}/${dash.name}`),
          }))}
        onSelectMenuOption={(dashboard) =>
          goto(`/${organization}/${project}/${dashboard}`)}
      />
    {/if}
  </ol>
</nav>
