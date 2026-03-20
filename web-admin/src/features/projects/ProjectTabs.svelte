<script lang="ts">
  import {
    position,
    width,
  } from "@rilldata/web-admin//components/nav/Tab.svelte";
  import Tab from "@rilldata/web-admin/components/nav/Tab.svelte";
  import { removeBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { type V1ProjectPermissions } from "../../client";

  export let projectPermissions: V1ProjectPermissions;
  export let organization: string;
  export let project: string;
  export let pathname: string;
  export let branchPrefix: string = "";

  const { chat, reports, alerts } = featureFlags;

  $: tabs = [
    {
      route: `/${organization}/${project}${branchPrefix}`,
      label: "Home",
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/ai`,
      label: "AI",
      hasPermission: $chat,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/dashboards`,
      label: "Dashboards",
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/query`,
      label: "Query",
      hasPermission: false,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/reports`,
      label: "Reports",
      hasPermission: $reports,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/alerts`,
      label: "Alerts",
      hasPermission: $alerts,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/status`,
      label: "Status",
      hasPermission: projectPermissions.manageProject,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/settings`,
      label: "Settings",
      hasPermission: projectPermissions.manageProject,
    },
  ];

  $: selectedIndex = tabs?.findLastIndex((t) => isSelected(t.route, pathname));

  function isSelected(tabRoute: string, currentPathname: string) {
    // Strip @branch from both sides so comparison works regardless of branch
    const normalizedTab = removeBranchFromPath(tabRoute);
    const normalizedPath = removeBranchFromPath(currentPathname);

    if (normalizedTab.endsWith(`/${organization}/${project}`)) {
      // For the Home (root) route, only exact match
      return normalizedPath === normalizedTab;
    }

    const isExactMatch = normalizedPath === normalizedTab;
    const isSubpage = normalizedPath.startsWith(normalizedTab + "/");

    return isExactMatch || isSubpage;
  }
</script>

<div class="bg-surface-base">
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
