<!-- ORG SETTINGS -->

<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/stores";
  import { V1BillingPlanType } from "@rilldata/web-admin/client";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import type { PageData } from "./$types";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";

  let {
    data,
    children,
  }: {
    data: PageData;
    children: Snippet;
  } = $props();

  let organization = $derived($page.params.organization);
  let basePage = $derived(`/${organization}/-/settings`);

  let planType = $derived(data.subscription?.plan?.planType);
  let planName = $derived(data.subscription?.plan?.name ?? "");
  let isEnterprise = $derived(
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
      isEnterprisePlan(planName),
  );

  let navItems = $derived([
    { label: "General", route: "", hasPermission: true },
    {
      label: "Billing",
      route: "/billing",
      hasPermission: true,
    },
    {
      label: "Usage",
      route: "/billing/usage",
      hasPermission: !isEnterprise,
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
