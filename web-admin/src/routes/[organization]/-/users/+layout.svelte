<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  // import {
  //   isOrgAdmin,
  //   isOrgEditor,
  // } from "@rilldata/web-admin/features/organizations/users/permissions";

  export let data;

  $: ({ organizationPermissions } = data);

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/users`;

  // $: isAdmin = isOrgAdmin(organizationPermissions);
  // $: isEditor = isOrgEditor(organizationPermissions);

  const navItems = [
    {
      label: "Users",
      route: "",
      hasPermission: true,
    },
    {
      label: "Groups",
      route: "/groups",
      // TODO: only org admin and editor can see this
      // hasPermission: isAdmin || isEditor,
      hasPermission: true,
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
