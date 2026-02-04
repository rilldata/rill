<!-- PROJECT SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: basePage = `/${organization}/${project}/-/settings`;

  const navItems = [
    {
      label: "Environment Variables",
      route: "/environment-variables",
      hasPermission: true,
    },
    {
      label: "Public URLs",
      route: "/public-urls",
      hasPermission: true,
    },
  ];
</script>

<ContentContainer title="Project settings" maxWidth={1100}>
  <div class="settings-layout">
    <aside class="nav-sidebar">
      <LeftNav
        {basePage}
        baseRoute="/[organization]/[project]/-/settings"
        {navItems}
        minWidth="180px"
      />
    </aside>
    <div class="content-area">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .settings-layout {
    @apply flex flex-col md:flex-row pt-6 gap-6 max-w-full flex-1;
  }

  .nav-sidebar {
    @apply shrink-0;
    @apply md:sticky md:top-0 md:self-start;
  }

  .content-area {
    @apply flex flex-col w-full min-w-0;
  }
</style>
