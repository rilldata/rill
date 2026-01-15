<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import LocalProjectStatusIndicator from "./LocalProjectStatusIndicator.svelte";

  $: currentPath = $page.url.pathname;

  $: activeTab = currentPath.includes("/ai")
    ? "ai"
    : currentPath.includes("/home")
      ? "home"
      : currentPath.includes("/edit")
        ? "edit"
        : currentPath.includes("/status")
          ? "status"
          : currentPath.includes("/settings")
            ? "settings"
            : "preview";
  const toolTip = "Deploy your project to access this feature";
  const tabs: {
    id: string;
    label: string;
    path: string;
    enabled?: boolean;
    disabledTooltip?: string;
  }[] = [
    { id: "home", label: "Home", path: "/home" },
    { id: "ai", label: "AI", path: "/ai" },
    { id: "preview", label: "Dashboards", path: "/preview" },
    { id: "Reports", label: "Reports", path: "/reports", enabled: false, disabledTooltip: toolTip},
    { id: "Alerts", label: "Alerts", path: "/alerts", enabled: false, disabledTooltip: toolTip },
    // { id: "edit", label: "Edit", path: "/edit" },
    { id: "status", label: "Status", path: "/status" },
    { id: "settings", label: "Settings", path: "/settings", enabled: false, disabledTooltip: toolTip },
  ];

  function isEnabled(tab: typeof tabs[number]): boolean {
    return tab.enabled !== false;
  }

  async function navigateTo(tab: typeof tabs[number]) {
    if (!isEnabled(tab)) return;
    await goto(tab.path);
  }
</script>

<div class="border-b border-gray-200 dark:border-gray-800 px-6 bg-white dark:bg-gray-950">
  <div class="flex gap-0">
    {#each tabs as tab (tab.id)}
      <div class="relative group">
        <button
          on:click={() => navigateTo(tab)}
          disabled={!isEnabled(tab)}
          class="px-4 py-3 text-sm font-medium transition-colors relative flex items-center gap-x-1
            {!isEnabled(tab)
              ? 'text-gray-400 dark:text-gray-600'
              : activeTab === tab.id
                ? 'text-gray-900 dark:text-white'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'}"
        >
          {tab.label}
          {#if tab.id === "status"}
            <LocalProjectStatusIndicator />
          {/if}
          {#if activeTab === tab.id && isEnabled(tab)}
            <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600" />
          {/if}
        </button>
        {#if !isEnabled(tab) && tab.disabledTooltip}
          <div class="absolute left-1/2 -translate-x-1/2 top-full mt-1 px-2 py-1 text-xs text-white bg-gray-800 dark:bg-gray-700 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
            {tab.disabledTooltip}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>
