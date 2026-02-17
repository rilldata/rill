<script lang="ts">
  import { page } from "$app/stores";
  import LocalProjectStatusIndicator from "./LocalProjectStatusIndicator.svelte";

  $: currentPath = $page.url.pathname;

  $: activeTab = currentPath.includes("/ai")
    ? "ai"
    : currentPath.includes("/home")
      ? "home"
      : currentPath.includes("/reports")
        ? "reports"
        : currentPath.includes("/alerts")
          ? "alerts"
          : currentPath.includes("/status")
            ? "status"
            : currentPath.includes("/settings")
              ? "settings"
              : "preview";

  const tabs: {
    id: string;
    label: string;
    path: string;
  }[] = [
    { id: "home", label: "Home", path: "/home" },
    { id: "ai", label: "AI", path: "/ai" },
    { id: "preview", label: "Dashboards", path: "/preview" },
    { id: "reports", label: "Reports", path: "/reports" },
    { id: "alerts", label: "Alerts", path: "/alerts" },
    { id: "status", label: "Status", path: "/status" },
    { id: "settings", label: "Settings", path: "/settings" },
  ];

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
    updateIndicator();
  }
</script>

<svelte:window on:resize={updateIndicator} />

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
    @apply bg-gray-100;
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
    transition: transform 0.2s ease, width 0.2s ease;
  }
</style>
