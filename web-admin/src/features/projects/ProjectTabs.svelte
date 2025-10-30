<script lang="ts">
  import {
    position,
    width,
  } from "@rilldata/web-admin//components/nav/Tab.svelte";
  import Tab from "@rilldata/web-admin/components/nav/Tab.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { type V1ProjectPermissions } from "../../client";

  export let projectPermissions: V1ProjectPermissions;
  export let organization: string;
  export let project: string;
  export let pathname: string;

  const { chat, reports, alerts } = featureFlags;

  $: tabs = [
    {
      route: `/${organization}/${project}`,
      label: "Home",
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}/-/ai`,
      label: "AI",
      hasPermission: $chat,
    },
    {
      route: `/${organization}/${project}/-/dashboards`,
      label: "Dashboards",
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}/-/reports`,
      label: "Reports",
      hasPermission: $reports,
    },
    {
      route: `/${organization}/${project}/-/alerts`,
      label: "Alerts",
      hasPermission: $alerts,
    },
    {
      route: `/${organization}/${project}/-/status`,
      label: "Status",
      hasPermission: projectPermissions.manageProject,
    },
    {
      // TODO: Change this back to `/${organization}/${project}/-/settings`
      // Once project settings are implemented
      route: `/${organization}/${project}/-/settings/environment-variables`,
      label: "Settings",
      hasPermission: projectPermissions.manageProject,
    },
  ];

  $: selectedIndex = tabs?.findLastIndex((t) => isSelected(t.route, pathname));

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

<div>
  <nav>
    {#each tabs as tab, i (tab.route)}
      {#if tab.hasPermission}
        <Tab
          route={tab.route}
          label={tab.label}
          selected={selectedIndex === i}
          {organization}
          {project}
        />
      {/if}
    {/each}
  </nav>

  {#if $width && $position}
    <span
      style:width="{$width}px"
      style:transform="translateX({$position}px) "
    />
  {/if}
</div>

<style lang="postcss">
  div {
    @apply border-b pt-1;
    @apply gap-y-[3px] flex flex-col;
  }

  nav {
    @apply flex w-fit;
    @apply gap-x-3 px-[17px];
  }

  span {
    @apply h-[3px] bg-primary-500 rounded transition-all;
  }
</style>
