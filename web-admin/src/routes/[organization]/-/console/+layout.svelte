<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  let { children }: { children: Snippet } = $props();

  let organization = $derived($page.params.organization);
  let basePage = $derived(`/${organization}/-/console`);

  let navItems = $derived([
    { label: "Overview", route: "", hasPermission: true },
    { label: "Projects", route: "/projects", hasPermission: true },
    { label: "Resources", route: "/resources", hasPermission: true },
  ]);
</script>

<svelte:head>
  <title>Admin Console - Rill</title>
</svelte:head>

<ContentContainer title="Admin Console" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/-/console"
      {navItems}
      minWidth="180px"
    />
    <div class="flex flex-col gap-y-6 w-full">
      {@render children()}
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
