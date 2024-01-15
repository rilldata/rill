<script lang="ts">
  import { afterNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "../../client";
  import Tab from "../../components/tabs/Tab.svelte";
  import TabGroup from "../../components/tabs/TabGroup.svelte";
  import TabList from "../../components/tabs/TabList.svelte";
  import ProjectDeploymentStatusChip from "./status/ProjectDeploymentStatusChip.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  // Get the list of tabs to display, depending on the user's permissions
  $: tabsQuery = createAdminServiceGetProject(organization, project, {
    query: {
      select: (data) => {
        const commonTabs = [
          {
            route: `/${organization}/${project}`,
            label: "Dashboards",
          },
          {
            route: `/${organization}/${project}/-/reports`,
            label: "Reports",
          },
        ];
        const adminTabs = [
          {
            route: `/${organization}/${project}/-/status`,
            label: "Status",
          },
        ];

        if (data.projectPermissions?.manageProject) {
          return [...commonTabs, ...adminTabs];
        } else {
          return commonTabs;
        }
      },
    },
  });
  $: tabs = $tabsQuery.data;

  function getCurrentTabIndex(tabs: { route: string }[], pathname: string) {
    return tabs.findIndex((tab) => {
      return tab.route === pathname;
    });
  }
  $: currentTabIndex = tabs && getCurrentTabIndex(tabs, $page.url.pathname);

  function handleTabChange(event: CustomEvent) {
    // Navigate to the new tab
    goto(`${tabs[event.detail].route}`);
  }

  afterNavigate((nav) => {
    // If changing to a new project, switch to the dashboards tab
    if (nav.from?.params && nav.to.params.project !== nav.from.params.project) {
      // We use DOM manipulation here because the library does not support controlled tabs
      // See: https://github.com/rgossiaux/svelte-headlessui/issues/80
      const dashboardTab = Array.from(
        document.querySelectorAll('button[role="tab"]'),
      ).find(
        (el) => (el as HTMLElement).innerText === "Dashboards",
      ) as HTMLButtonElement;
      dashboardTab.click();
    }
  });
</script>

{#if tabs}
  <div class="pl-[17px] border-b pt-1 pb-[3px]">
    <TabGroup defaultIndex={currentTabIndex} on:change={handleTabChange}>
      <TabList>
        {#each tabs as tab}
          <Tab>
            {tab.label}
            {#if tab.label === "Status"}
              <ProjectDeploymentStatusChip
                {organization}
                {project}
                iconOnly={true}
              />
            {/if}
          </Tab>
        {/each}
      </TabList>
    </TabGroup>
  </div>
{/if}
