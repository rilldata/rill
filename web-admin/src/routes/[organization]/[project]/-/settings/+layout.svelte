<!-- PROJECT SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: basePage = `/${organization}/${project}/-/settings`;

  const navItems = [
    {
      label: "Environment Variables",
      route: "/environment-variables",
    },
    {
      label: "Public URLs",
      route: "/public-urls",
    },
  ];
</script>

<div class="layout-container">
  <h3>Settings</h3>

  <div class="container">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/[project]/-/settings"
      {navItems}
      minWidth="180px"
    />
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
