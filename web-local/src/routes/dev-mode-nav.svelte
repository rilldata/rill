<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";

  $: currentPath = $page.url.pathname;

  $: activeTab = currentPath.includes("/edit")
    ? "edit"
    : currentPath.includes("/status")
      ? "status"
      : currentPath.includes("/settings")
        ? "settings"
        : "preview";

  const tabs = [
    { id: "preview", label: "Preview", path: "/preview" },
    { id: "edit", label: "Edit", path: "/edit" },
    { id: "status", label: "Status", path: "/status" },
    { id: "settings", label: "Settings", path: "/settings" },
  ];

  async function navigateTo(path: string) {
    await goto(path);
  }
</script>

<div class="border-b border-gray-200 dark:border-gray-800 px-6 bg-white dark:bg-gray-950">
  <div class="flex gap-0">
    {#each tabs as tab (tab.id)}
      <button
        on:click={() => navigateTo(tab.path)}
        class="px-4 py-3 text-sm font-medium transition-colors relative {activeTab === tab.id
          ? 'text-gray-900 dark:text-white'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'}"
      >
        {tab.label}
        {#if activeTab === tab.id}
          <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600" />
        {/if}
      </button>
    {/each}
  </div>
</div>
