<!-- web-admin/src/features/admin/layout/AdminSidebar.svelte -->
<script lang="ts">
  import { page } from "$app/stores";

  const navGroups = [
    {
      heading: "People",
      items: [
        { label: "Users", href: "/-/admin" },
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
      ],
    },
  ];

  function isActive(href: string, pathname: string): boolean {
    if (href === "/-/admin") return pathname === "/-/admin";
    return pathname.startsWith(href);
  }
</script>

<nav
  class="w-56 flex-shrink-0 border-r border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 flex flex-col h-full"
>
  <div class="px-4 py-4 border-b border-slate-200 dark:border-slate-700">
    <span class="text-sm font-semibold text-slate-900 dark:text-slate-100">
      Admin Console
    </span>
  </div>

  <div class="flex-1 overflow-y-auto py-3 px-3">
    {#each navGroups as group}
      <div class="mb-4">
        <span
          class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 px-2 mb-1 block"
        >
          {group.heading}
        </span>
        {#each group.items as item}
          <a
            href={item.href}
            class="block px-2 py-1.5 text-sm rounded-md transition-colors {isActive(
              item.href,
              $page.url.pathname,
            )
              ? 'bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-slate-100 font-medium'
              : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'}"
          >
            {item.label}
          </a>
        {/each}
      </div>
    {/each}
  </div>
</nav>
