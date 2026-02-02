<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetOrganization,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { isEnterprisePlan, isManagedPlan } from "./plans/utils";

  export let organization: string;
  export let subscription: V1Subscription;

  $: org = createAdminServiceGetOrganization(organization);
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;
  $: neverSubscribed = $categorisedIssues.data?.neverSubscribed;
  $: onTrial = !!$categorisedIssues.data?.trial;
  $: onEnterprisePlan =
    subscription?.plan && isEnterprisePlan(subscription.plan.name);
  $: onManagedPlan =
    subscription?.plan && isManagedPlan(subscription.plan.name);
  $: hidePaymentModule =
    neverSubscribed || onTrial || onEnterprisePlan || onManagedPlan;

  function handleManagePayment() {
    goto(`/${organization}/-/settings/billing/payment`);
  }
</script>

<!-- Presence of paymentCustomerId signifies that the org's payment is managed through stripe -->
{#if !$categorisedIssues.isLoading && $org.data?.organization?.paymentCustomerId && !hidePaymentModule}
  <SettingsContainer title="Payment Method">
    <div slot="body" class="flex flex-row items-center gap-x-1">
      {#if paymentIssues?.length}
        <CancelCircle className="text-red-600" size="14px" />
        {getPaymentIssueErrorText(paymentIssues)} Please click Manage below to correct.
      {:else}
        Your payment method is valid and good to go.
      {/if}
    </div>
    <Button slot="action" type="secondary" onClick={handleManagePayment}>
      Manage
    </Button>
  </SettingsContainer>
{/if}
