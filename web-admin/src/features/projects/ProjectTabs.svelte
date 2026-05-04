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
  /** When rendered inside the cloud editor's dev-preview chrome, route the
   *  tabs to `/-/edit/{section}` so navigation stays inside the editor
   *  session instead of dropping the user into the production project view. */
  export let editMode: boolean = false;

  const { chat, reports, alerts } = featureFlags;

  $: base = `/${organization}/${project}${branchPrefix}`;
  $: sectionPrefix = editMode ? `${base}/-/edit` : `${base}/-`;

  $: tabs = [
    {
      route: base,
      label: "Home",
      // In dev-preview chrome there's no separate project-home page —
      // Home would collide with Dashboards (`/-/edit/dashboards`), so
      // drop it.
      hasPermission: !editMode,
    },
    {
      route: `${sectionPrefix}/ai`,
      label: "AI",
      hasPermission: $chat,
    },
    {
      route: `${sectionPrefix}/dashboards`,
      label: "Dashboards",
      hasPermission: true,
    },
    {
      route: `${sectionPrefix}/query`,
      label: "Query",
      hasPermission: false,
    },
    {
      route: `${sectionPrefix}/reports`,
      label: "Reports",
      hasPermission: $reports,
    },
    {
      route: `${sectionPrefix}/alerts`,
      label: "Alerts",
      hasPermission: $alerts,
    },
    {
      route: `${sectionPrefix}/status`,
      label: "Status",
      hasPermission: projectPermissions.manageProject,
    },
    {
      route: `${sectionPrefix}/settings`,
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

  {#if $width}
    <span style:width="{$width}px" style:transform="translateX({$position}px) "
    ></span>
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
