<script lang="ts">
  import {
    position,
    width,
  } from "@rilldata/web-admin//components/nav/Tab.svelte";
  import Tab from "@rilldata/web-admin/components/nav/Tab.svelte";
  import { removeBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onMount } from "svelte";
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
      label: m.nav_tab_home(),
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/ai`,
      label: m.nav_tab_ai(),
      hasPermission: $chat,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/dashboards`,
      label: m.nav_tab_dashboards(),
      hasPermission: true,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/query`,
      label: m.nav_tab_query(),
      hasPermission: false,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/reports`,
      label: m.nav_tab_reports(),
      hasPermission: $reports,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/alerts`,
      label: m.nav_tab_alerts(),
      hasPermission: $alerts,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/status`,
      label: m.nav_tab_status(),
      hasPermission: projectPermissions.manageProject,
    },
    {
      route: `/${organization}/${project}${branchPrefix}/-/settings`,
      label: m.nav_tab_settings(),
      hasPermission: projectPermissions.manageProject,
    },
  ];

  $: selectedIndex = tabs?.findLastIndex((t) => isSelected(t.route, pathname));

  let scroller: HTMLDivElement;
  let navElement: HTMLElement;
  let canScrollLeft = false;
  let canScrollRight = false;

  function updateFades() {
    if (!scroller) return;
    canScrollLeft = scroller.scrollLeft > 0;
    canScrollRight =
      scroller.scrollLeft + scroller.clientWidth < scroller.scrollWidth - 1;
  }

  onMount(() => {
    updateFades();
    // Observe the nav too: tabs mount/unmount with permissions and feature
    // flags, which changes scrollWidth without resizing the scroller itself.
    const observer = new ResizeObserver(updateFades);
    observer.observe(scroller);
    observer.observe(navElement);
    return () => observer.disconnect();
  });

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

<div class="tabs-container bg-surface-base">
  <!-- On narrow viewports the tab row scrolls horizontally; the selection
       indicator scrolls with it since both share the scroller. The edge
       fades signal the hidden overflow on either side. -->
  <div class="tabs-scroller" bind:this={scroller} on:scroll={updateFades}>
    <nav bind:this={navElement}>
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
      <span
        style:width="{$width}px"
        style:transform="translateX({$position}px) "
      ></span>
    {/if}
  </div>

  <div class="fade left" class:visible={canScrollLeft}></div>
  <div class="fade right" class:visible={canScrollRight}></div>
</div>

<style lang="postcss">
  .tabs-container {
    @apply relative border-b;
  }

  .tabs-scroller {
    @apply pt-1;
    @apply gap-y-[3px] flex flex-col;
    @apply overflow-x-auto;
  }

  .fade {
    @apply absolute inset-y-0 w-8 pointer-events-none;
    @apply opacity-0 transition-opacity duration-200;
  }

  .fade.visible {
    @apply opacity-100;
  }

  .fade.left {
    @apply left-0 bg-gradient-to-r from-surface-base to-transparent;
  }

  .fade.right {
    @apply right-0 bg-gradient-to-l from-surface-base to-transparent;
  }

  nav {
    @apply flex w-fit;
    @apply gap-x-3 px-[17px];
  }

  span {
    @apply h-[3px] bg-primary-500 rounded transition-all;
  }
</style>
