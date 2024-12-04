<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "../../client";
  import ProjectGlobalStatusIndicator from "./status/ProjectGlobalStatusIndicator.svelte";

  $: ({
    url: { pathname },
    params: { organization, project },
  } = $page);

  // Get the list of tabs to display, depending on the user's permissions
  $: tabsQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    {
      query: {
        select: (data) => {
          let commonTabs = [
            {
              route: `/${organization}/${project}`,
              label: "Dashboards",
            },
            {
              route: `/${organization}/${project}/-/reports`,
              label: "Reports",
            },
          ];

          commonTabs.push({
            route: `/${organization}/${project}/-/alerts`,
            label: "Alerts",
          });

          const adminTabs = [
            {
              route: `/${organization}/${project}/-/status`,
              label: "Status",
            },
            {
              // TODO: Change this back to `/${organization}/${project}/-/settings`
              // Once project settings are implemented
              route: `/${organization}/${project}/-/settings/environment-variables`,
              label: "Settings",
            },
          ];

          if (data.projectPermissions?.manageProject) {
            return [...commonTabs, ...adminTabs];
          } else {
            return commonTabs;
          }
        },
      },
    },
  );

  $: tabs = $tabsQuery.data;

  function isSelected(tabRoute: string, currentPathname: string) {
    if (tabRoute.endsWith(`/${organization}/${project}`)) {
      // For the dashboard (root) route, only exact match
      return currentPathname === tabRoute;
    }

    const isExactMatch = currentPathname === tabRoute;
    const isSubpage = currentPathname.startsWith(tabRoute + "/");

    return isExactMatch || isSubpage;
  }
</script>

{#if tabs}
  <nav>
    {#each tabs as tab (tab.route)}
      <a href={tab.route} class:selected={isSelected(tab.route, pathname)}>
        {tab.label}
        {#if tab.label === "Status"}
          <ProjectGlobalStatusIndicator {organization} {project} />
        {/if}
      </a>
    {/each}
  </nav>
{/if}

<style lang="postcss">
  a {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm text-gray-500;
    @apply text-xs font-medium justify-center;
  }

  .selected {
    @apply text-gray-900;
  }

  a:hover {
    @apply bg-slate-100 text-gray-700;
  }

  nav {
    @apply flex gap-x-6 px-[17px] border-b pt-1 pb-[3px];
  }
</style>
