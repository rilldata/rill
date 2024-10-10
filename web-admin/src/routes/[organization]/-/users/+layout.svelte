<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/users`;

  const navItems = [
    {
      label: "Users",
      route: "",
    },
    // TODO: Uncomment this once we have user groups
    // {
    //   label: "Groups",
    //   route: "/groups",
    // },
  ];
</script>

<div class="layout-container">
  <h3>Users</h3>
  <div class="container">
    <LeftNav {basePage} baseRoute="/[organization]/-/users" {navItems} />
    <slot />
  </div>
</div>

<style lang="postcss">
  .layout-container {
    @apply px-32 py-10;
  }

  h3 {
    @apply text-2xl font-semibold;
  }

  .container {
    @apply flex flex-row pt-6 gap-x-6 max-w-full;
  }
</style>
