<!-- PROJECT STATUS -->

<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/${$page.params.project}/-/status`;

  $: navItems = [
    {
      label: "Overview",
      route: "",
      hasPermission: true,
    },
    {
      label: "Deployments",
      route: "/branches",
      hasPermission: true,
    },
    {
      label: "Resources",
      route: "/resources",
      hasPermission: true,
    },
    {
      label: "Tables",
      route: "/tables",
      hasPermission: true,
    },
    {
      label: "Logs",
      route: "/logs",
      hasPermission: true,
    },
    {
      label: "Analytics",
      route: "/analytics",
      hasPermission: false,
    },
  ];
</script>

<ContentContainer title="Project Status" maxWidth={1100}>
  <div class="container flex-col lg:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/[project]/-/status"
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
