<!-- ORG SETTINGS -->

<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import type { PageData } from "./$types";
  import { PaidPlanTypes } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { children, data }: { children: Snippet; data: PageData } = $props();

  let { billingPortalUrl } = $derived(data);
  let showUsageSettings = $derived(Boolean(billingPortalUrl));

  let organization = $derived($page.params.organization);
  let basePage = $derived(`/${organization}/-/settings`);

  let isPaidPlan = $derived(
    data.subscription?.plan?.planType &&
      PaidPlanTypes[data.subscription.plan.planType],
  );

  // The Usage tab is intentionally hidden for all plans until the new usage
  // page is ready. Pro and Team users still get a `View detailed usage` link
  // out to the Orb billing portal from the Plan card.
  let navItems = $derived([
    { label: m.settings_nav_general(), route: "", hasPermission: true },
    { label: m.settings_nav_billing(), route: "/billing", hasPermission: true },
    ...(isPaidPlan
      ? [
          {
            label: m.settings_nav_usage(),
            route: "/usage",
            hasPermission: showUsageSettings,
          },
        ]
      : []),
  ]);
</script>

<ContentContainer title={m.settings_org_page_title()} maxWidth={1100}>
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
