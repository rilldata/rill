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
  <div class="container flex-col md:flex-row">
    <div class="nav-wrapper">
      <LeftNav
        {basePage}
        baseRoute="/[organization]/[project]/-/settings"
        {navItems}
        minWidth="180px"
      />
    </div>
    <div class="flex flex-col w-full min-w-0">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full items-start flex-1;
  }

  .nav-wrapper {
    @apply md:sticky md:top-0 shrink-0;
  }
</style>
