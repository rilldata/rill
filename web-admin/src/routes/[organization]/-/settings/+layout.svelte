<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/settings`;

  const navItems = [
    {
      label: "General",
      route: "general",
    },
    {
      label: "Billing",
      route: "billing",
    },
    {
      label: "Usage",
      route: "usage",
    },
  ];
</script>

<LeftNav
  title="Settings"
  {basePage}
  baseRoute="/[organization]/-/settings"
  {navItems}
>
  <slot />
</LeftNav>
