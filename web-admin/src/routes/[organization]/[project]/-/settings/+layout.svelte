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
    },
    {
      label: "Public URLs",
      route: "/public-urls",
    },
  ];
</script>

<ContentContainer title="Project settings" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/[project]/-/settings"
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
