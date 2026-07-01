<!-- ORG SETTINGS -->

<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import type { PageData } from "./$types";

  let { children, data }: { children: Snippet; data: PageData } = $props();

  let { billingPortalUrl } = $derived(data);
  let showUsageSettings = $derived(Boolean(billingPortalUrl));

  let organization = $derived($page.params.organization);
  let basePage = $derived(`/${organization}/-/settings`);

  // The Usage tab is intentionally hidden for all plans until the new usage
  // page is ready. Pro and Team users still get a `View detailed usage` link
  // out to the Orb billing portal from the Plan card.
  let navItems = $derived([
    { label: "General", route: "", hasPermission: true },
    { label: "Billing", route: "/billing", hasPermission: true },
    { label: "Usage", route: "/usage", hasPermission: showUsageSettings },
  ]);
</script>

<ContentContainer title="Organization settings" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/-/settings"
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
