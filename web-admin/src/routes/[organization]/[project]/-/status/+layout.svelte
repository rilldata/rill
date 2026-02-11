<!-- PROJECT STATUS -->

<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  const navItems = [
    {
      label: "Status",
      route: "/project-status",
      hasPermission: true,
    },
    {
      label: "Resources",
      route: "/resources",
      hasPermission: true,
    },
    {
      label: "Tables",
      route: "/model-details",
      hasPermission: true,
    },
    {
      label: "Logs",
      route: "/project-logs",
      hasPermission: true,
    },
    {
      label: "Analytics",
      route: "/analytics",
      hasPermission: true,
    },
  ];
</script>

<ContentContainer title="Project Configuration" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
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
