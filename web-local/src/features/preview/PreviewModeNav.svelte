<script lang="ts">
  import { page } from "$app/stores";
  import { tick } from "svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import LocalProjectStatusIndicator from "../../routes/LocalProjectStatusIndicator.svelte";

  const { chat } = featureFlags;

  $: currentPath = $page.url.pathname;

  const baseTabs = [
    { id: "dashboards", label: "Dashboards", path: "/dashboards" },
    { id: "status", label: "Status", path: "/status" },
  ];

  const aiTab = { id: "ai", label: "AI", path: "/ai" };

  $: tabs = $chat ? [baseTabs[0], aiTab, baseTabs[1]] : baseTabs;

  $: activeTab =
    tabs.find((t) => currentPath.startsWith(t.path))?.id ?? "dashboards";

  $: selectedIndex = tabs.findIndex((t) => t.id === activeTab);

  // Track tab element positions for the animated underline
  let tabElements: HTMLAnchorElement[] = [];
  let indicatorLeft = 0;
  let indicatorWidth = 0;

  function updateIndicator() {
    const el = tabElements[selectedIndex];
    if (el) {
      indicatorLeft = el.offsetLeft;
      indicatorWidth = el.clientWidth;
    }
  }

  $: if (selectedIndex >= 0 && tabElements.length) {
    tick().then(updateIndicator);
  }
</script>

<svelte:window onresize={updateIndicator} />

<div class="nav-bar">
  <nav>
    {#each tabs as tab, i (tab.id)}
      <a
        href={tab.path}
        class="tab"
        class:selected={activeTab === tab.id}
        bind:this={tabElements[i]}
      >
        <p data-content={tab.label}>
          {tab.label}
        </p>
        {#if tab.id === "status"}
          <LocalProjectStatusIndicator />
        {/if}
      </a>
    {/each}
  </nav>

  {#if indicatorWidth > 0}
    <span
      class="indicator"
      style:width="{indicatorWidth}px"
      style:transform="translateX({indicatorLeft}px)"
    />
  {/if}
</div>

<style lang="postcss">
  .nav-bar {
    @apply bg-surface-base border-b pt-1;
    @apply gap-y-[3px] flex flex-col;
  }

  nav {
    @apply flex w-fit;
    @apply gap-x-3 px-[17px];
  }

  .tab {
    @apply px-2 py-1.5 flex gap-x-1 items-center w-fit;
    @apply rounded-sm text-fg-muted;
    @apply text-xs font-medium justify-center;
    @apply no-underline;
  }

  .selected {
    @apply text-fg-accent font-semibold;
  }

  .tab:hover {
    background: var(--surface-hover);
  }

  .tab p {
    @apply text-center;
  }

  /* Prevent layout shift on font weight change */
  .tab p::before {
    content: attr(data-content);
    display: block;
    font-weight: 600;
    height: 0;
    visibility: hidden;
  }

  .indicator {
    @apply h-[3px] bg-primary-500 rounded;
    transition:
      transform 0.2s ease,
      width 0.2s ease;
  }
</style>
