<script lang="ts">
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  const basePage = "/-/superuser";
  const baseRoute = "/-/superuser";

  const navGroups = [
    {
      heading: "People",
      items: [
        { label: "Users", route: "" },
        { label: "Superusers", route: "/superusers" },
      ],
    },
    {
      heading: "Billing & Plans",
      items: [
        { label: "Billing", route: "/billing" },
        { label: "Quotas", route: "/quotas" },
      ],
    },
    {
      heading: "Resources",
      items: [
        { label: "Organizations", route: "/organizations" },
        { label: "Projects", route: "/projects" },
      ],
    },
  ];
</script>

<nav class="w-56 flex-shrink-0 border-r flex flex-col h-full">
  <div class="px-4 py-4 border-b">
    <span class="text-sm font-semibold text-fg-primary">
      Superuser Console
    </span>
  </div>

  <div class="flex-1 overflow-y-auto py-3 px-3">
    {#each navGroups as group}
      <div class="mb-4">
        <span
          class="text-sm font-semibold uppercase tracking-wider text-fg-muted px-2 mb-1 block"
        >
          {group.heading}
        </span>
        <LeftNav
          {basePage}
          {baseRoute}
          navItems={group.items}
          minWidth="auto"
        />
      </div>
    {/each}
  </div>
</nav>
