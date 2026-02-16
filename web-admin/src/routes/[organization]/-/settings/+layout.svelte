<script lang="ts">
  import { page } from "$app/stores";
  import SettingsNav from "@rilldata/web-admin/src/lib/components/navigation/SettingsNav.svelte";

  $: organization = $page.params.organization;

  const settingsNavItems = [
    {
      label: "General",
      href: `/${organization}/-/settings`,
      exactMatch: true,
    },
    {
      label: "Members",
      href: `/${organization}/-/settings/members`,
    },
    {
      label: "Billing",
      href: `/${organization}/-/settings/billing`,
    },
    {
      label: "Tokens",
      href: `/${organization}/-/settings/tokens`,
    },
  ];

  // Reactive: update hrefs when organization changes
  $: navItems = settingsNavItems.map((item) => ({
    ...item,
    href: item.href.replace(
      settingsNavItems[0].href.split("/-/settings")[0],
      `/${organization}`,
    ),
  }));
</script>

<div class="flex flex-col gap-5 p-5 w-full max-w-[900px] mx-auto">
  <h1 class="text-2xl font-bold">Settings</h1>
  <div class="flex gap-8">
    <nav class="flex flex-col gap-1 min-w-[160px] shrink-0">
      {#each navItems as item}
        {@const isActive = item.exactMatch
          ? $page.url.pathname === item.href ||
            $page.url.pathname === item.href + "/"
          : $page.url.pathname.startsWith(item.href)}
        <a
          href={item.href}
          class="px-3 py-1.5 rounded-sm text-sm font-medium transition-colors
            {isActive
            ? 'bg-gray-100 text-gray-900'
            : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'}"
          aria-current={isActive ? "page" : undefined}
        >
          {item.label}
        </a>
      {/each}
    </nav>
    <div class="flex-1 min-w-0">
      <slot />
    </div>
  </div>
</div>