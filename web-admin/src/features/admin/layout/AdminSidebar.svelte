<!-- web-admin/src/features/admin/layout/AdminSidebar.svelte -->
<script lang="ts">
  import { page } from "$app/stores";

  const navGroups = [
    {
      heading: "Overview",
      items: [{ label: "Dashboard", href: "/-/admin" }],
    },
    {
      heading: "People",
      items: [
        { label: "Users", href: "/-/admin/users" },
        { label: "Superusers", href: "/-/admin/superusers" },
      ],
    },
    {
      heading: "Billing & Plans",
      items: [
        { label: "Billing", href: "/-/admin/billing" },
        { label: "Quotas", href: "/-/admin/quotas" },
      ],
    },
    {
      heading: "Resources",
      items: [
        { label: "Organizations", href: "/-/admin/organizations" },
        { label: "Projects", href: "/-/admin/projects" },
        { label: "Whitelist", href: "/-/admin/whitelist" },
      ],
    },
    {
      heading: "Advanced",
      items: [
        { label: "Annotations", href: "/-/admin/annotations" },
        { label: "Virtual Files", href: "/-/admin/virtual-files" },
        { label: "Runtime", href: "/-/admin/runtime" },
      ],
    },
  ];

  function isActive(href: string, pathname: string): boolean {
    if (href === "/-/admin") return pathname === "/-/admin";
    return pathname.startsWith(href);
  }
</script>

<nav class="sidebar">
  <div class="sidebar-header">
    <span class="logo-text">Admin Console</span>
  </div>

  <div class="sidebar-content">
    {#each navGroups as group}
      <div class="nav-group">
        <span class="group-heading">{group.heading}</span>
        {#each group.items as item}
          <a
            href={item.href}
            class="nav-item"
            class:active={isActive(item.href, $page.url.pathname)}
          >
            {item.label}
          </a>
        {/each}
      </div>
    {/each}
  </div>
</nav>

<style lang="postcss">
  .sidebar {
    @apply w-56 flex-shrink-0 border-r border-slate-200 dark:border-slate-700
      bg-white dark:bg-slate-900 flex flex-col h-full;
  }

  .sidebar-header {
    @apply px-4 py-4 border-b border-slate-200 dark:border-slate-700;
  }

  .logo-text {
    @apply text-sm font-semibold text-slate-900 dark:text-slate-100;
  }

  .sidebar-content {
    @apply flex-1 overflow-y-auto py-3 px-3;
  }

  .nav-group {
    @apply mb-4;
  }

  .group-heading {
    @apply text-[11px] font-semibold uppercase tracking-wider
      text-slate-400 dark:text-slate-500 px-2 mb-1 block;
  }

  .nav-item {
    @apply block px-2 py-1.5 text-sm rounded-md
      text-slate-600 dark:text-slate-300
      hover:bg-slate-100 dark:hover:bg-slate-800
      transition-colors;
  }

  .nav-item.active {
    @apply bg-slate-100 dark:bg-slate-800
      text-slate-900 dark:text-slate-100 font-medium;
  }
</style>
