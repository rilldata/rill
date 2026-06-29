<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
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
      label: m.users_tab_members({ count: membersCount }),
      route: "",
      hasPermission: hasManageOrgMembers,
    },
    {
      label: m.users_tab_guests({ count: guestsCount }),
      route: "/guests",
      hasPermission: hasManageOrgMembers,
    },
    {
      label: m.users_tab_groups({ count: groupsCount }),
      route: "/groups",
      hasPermission: hasManageOrgMembers,
    },
  ];
</script>

<ContentContainer title={m.users_page_title()} maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav {basePage} baseRoute="/[organization]/-/users" {navItems} />
    <slot />
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
