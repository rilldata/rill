<!-- ORG SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import type { PageData } from "./$types";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";

  export let data: PageData;

  $: ({ subscription, neverSubscribed, billingPortalUrl } = data);

  $: organization = $page.params.organization;
  $: basePage = `/${organization}/-/settings`;
  $: onEnterprisePlan =
    subscription?.plan && isEnterprisePlan(subscription?.plan);
  $: hideBillingSettings = neverSubscribed;
  $: hideUsageSettings = onEnterprisePlan || !billingPortalUrl;

  $: navItems = [
    { label: "General", route: "" },
    ...(hideBillingSettings
      ? []
      : [
          { label: "Billing", route: "/billing" },
          ...(hideUsageSettings ? [] : [{ label: "Usage", route: "/usage" }]),
        ]),
  ];
</script>

<ContentContainer title="Settings" maxWidth={960}>
  <div class="container flex-col sm:flex-row">
    <LeftNav {basePage} baseRoute="/[organization]/-/settings" {navItems} />
    <div class="flex flex-col gap-y-6">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
