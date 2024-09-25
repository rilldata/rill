<!-- ORG SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/settings`;

  const navItems = [
    {
      label: "General",
      route: "",
    },
    // TODO: Uncomment when we have billing and usage page
    // {
    //   label: "Billing",
    //   route: "/billing",
    // },
    // {
    //   label: "Usage",
    //   route: "/usage",
    // },
  ];
</script>

<div class="layout-container">
  <h3>Settings</h3>
  <div class="container">
    <LeftNav {basePage} baseRoute="/[organization]/-/settings" {navItems} />
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
    @apply flex flex-row pt-6 gap-x-6;
  }
</style>
