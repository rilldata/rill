<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";

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

<ContentContainer title="Users" maxWidth={1100}>
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
