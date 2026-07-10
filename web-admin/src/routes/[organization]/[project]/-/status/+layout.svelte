<!-- PROJECT STATUS -->

<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;
  const { cloudEditing } = featureFlags;

  $: navItems = [
    {
      label: m.status_nav_overview(),
      route: "",
      hasPermission: true,
    },
    ...($cloudEditing
      ? [
          {
            label: m.status_nav_branches(),
            route: "/branches",
            hasPermission: true,
          },
        ]
      : []),
    {
      label: m.status_nav_resources(),
      route: "/resources",
      hasPermission: true,
    },
    {
      label: m.status_nav_tables(),
      route: "/tables",
      hasPermission: true,
    },
    {
      label: m.status_nav_logs(),
      route: "/logs",
      hasPermission: true,
    },
    {
      label: m.status_nav_analytics(),
      route: "/analytics",
      hasPermission: false,
    },
  ];
</script>

<ContentContainer title={m.status_page_title()} maxWidth={1100}>
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
