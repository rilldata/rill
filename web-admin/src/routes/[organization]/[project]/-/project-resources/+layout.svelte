<!-- PROJECT RESOURCES -->

<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/project-resources`;

  const navItems = [
    {
      label: "Alerts",
      route: "/alerts",
      hasPermission: true,
    },
    {
      label: "Reports",
      route: "/reports",
      hasPermission: true,
    },
    {
      label: "Public URLs",
      route: "/public-urls",
      hasPermission: true,
    },
  ];
</script>

<ContentContainer title="Project resources" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/[project]/-/project-resources"
      {navItems}
      minWidth="180px"
    />
    <slot />
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
