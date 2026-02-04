<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import { getUserCounts } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";

  export let data;

  $: ({ organizationPermissions } = data);

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/users`;

  // https://docs.rilldata.com/guide/administration/users-and-access/roles-permissions#organization-level-permissions
  // org admin and editor can manage org members
  $: hasManageOrgMembers = organizationPermissions?.manageOrgMembers;

  $: userCountsQuery = getUserCounts(organization);
  $: ({ membersCount, guestsCount, groupsCount } = $userCountsQuery);

  $: navItems = [
    {
      label: `Members (${membersCount})`,
      route: "",
      hasPermission: hasManageOrgMembers,
    },
    {
      label: `Guests (${guestsCount})`,
      route: "/guests",
      hasPermission: hasManageOrgMembers,
    },
    {
      label: `Groups (${groupsCount})`,
      route: "/groups",
      hasPermission: hasManageOrgMembers,
    },
  ];
</script>

<ContentContainer title="Manage users" maxWidth={1100}>
  <div class="settings-layout">
    <aside class="nav-sidebar">
      <LeftNav {basePage} baseRoute="/[organization]/-/users" {navItems} />
    </aside>
    <div class="content-area">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .settings-layout {
    @apply flex flex-col pt-6 gap-6 max-w-full flex-1;
  }

  .nav-sidebar {
    @apply shrink-0;
  }

  .content-area {
    @apply flex flex-col w-full min-w-0;
  }

  @media (min-width: 768px) {
    .settings-layout {
      @apply flex-row;
    }

    .nav-sidebar {
      position: sticky;
      top: 0;
      align-self: flex-start;
    }
  }
</style>
