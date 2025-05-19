<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";

  export let data;

  $: ({ organizationPermissions } = data);

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/users`;

  // https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
  // org admin and editor can manage org members
  $: hasManageOrgMembers = organizationPermissions?.manageOrgMembers;

  const navItems = [
    {
      label: "Users",
      route: "",
      hasPermission: hasManageOrgMembers,
    },
    {
      label: "Groups",
      route: "/groups",
      hasPermission: hasManageOrgMembers,
    },
  ];
</script>

<ContentContainer title="Manage users" maxWidth={1100}>
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
