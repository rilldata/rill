<!-- ORG SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import type { PageData } from "./$types";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import type { Snippet } from "svelte";

  let {
    data,
    children,
  }: {
    data: PageData;
    children: Snippet;
  } = $props();

  let {
    subscription,
    neverSubscribed,
    billingPortalUrl,
    organizationPermissions,
  } = $derived(data);

  let organization = $derived($page.params.organization);
  let basePage = $derived(`/${organization}/-/settings`);
  let onEnterprisePlan = $derived(
    subscription?.plan?.name && isEnterprisePlan(subscription.plan.name),
  );
  let hideBillingSettings = $derived(neverSubscribed);
  let hideUsageSettings = $derived(onEnterprisePlan || !billingPortalUrl);
  let canManageOrg = $derived(!!organizationPermissions?.manageOrg);

  let navItems = $derived([
    { label: "General", route: "", hasPermission: true },
    {
      label: "Service Accounts",
      route: "/service-accounts",
      hasPermission: canManageOrg,
    },
    {
      label: "Billing",
      route: "/billing",
      hasPermission: !hideBillingSettings,
    },
    {
      label: "Usage",
      route: "/usage",
      hasPermission: !hideBillingSettings && !hideUsageSettings,
    },
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
