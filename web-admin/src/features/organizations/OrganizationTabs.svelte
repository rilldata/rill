<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";

  $: ({
    url: { pathname },
    params: { organization },
  } = $page);

  // Get the list of tabs to display, depending on the user's permissions
  $: tabsQuery = createAdminServiceGetOrganization(organization, {
    query: {
      select: (data) => {
        let tabs = [
          {
            route: `/${organization}`,
            label: "Projects",
          },
        ];

        if (data.permissions.manageOrgMembers) {
          tabs.push({
            route: `/${organization}/-/users`,
            label: "Users",
          });
        }

        if (data.permissions.manageOrg) {
          tabs.push({
            route: `/${organization}/-/settings`,
            label: "Settings",
          });
        }

        return tabs;
      },
    },
  });

  $: tabs = $tabsQuery.data;
  // 1st entry is always the default page. so findLastIndex will make sure the correct page is matched.
  $: selectedIndex = tabs?.findLastIndex((t) => pathname.startsWith(t.route));
</script>

<!-- Hide the tabs when there is only one entry -->
{#if tabs?.length && tabs?.length > 1}
  <nav>
    {#each tabs as tab, i (tab.route)}
      <a href={tab.route} class:selected={selectedIndex === i}>
        {tab.label}
      </a>
    {/each}
  </nav>
{:else}
  <!-- Add a border to keep things consistent. It is cleaner to handle this here. -->
  <div class="border-b"></div>
{/if}

<style lang="postcss">
  a {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm text-gray-500;
    @apply text-xs font-medium justify-center;
  }

  .selected {
    @apply text-gray-900;
  }

  a:hover {
    @apply bg-slate-100 text-gray-700;
  }

  nav {
    @apply flex gap-x-6 px-[17px] border-b pt-1 pb-[3px];
  }
</style>
