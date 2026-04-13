<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceGetOrganization,
  } from "@rilldata/web-admin/client";
  import {
    getPaymentIssueErrorText,
    needsPaymentSetup,
  } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { isEnterprisePlan, isManagedPlan } from "./plans/utils";

  let { organization }: { organization: string } = $props();

  let org = $derived(createAdminServiceGetOrganization(organization));
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let plan = $derived($subscriptionQuery?.data?.subscription?.plan);
  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let hasPaymentCustomer = $derived(
    !!$org.data?.organization?.paymentCustomerId,
  );
  let paymentIssues = $derived($categorisedIssues.data?.payment);
  let neverSubscribed = $derived(!!$categorisedIssues.data?.neverSubscribed);
  let onTrial = $derived(!!$categorisedIssues.data?.trial);
  let onManagedPlan = $derived(plan && isManagedPlan(plan.name));
  let onEnterprisePlan = $derived(plan && isEnterprisePlan(plan.name));
  let pendingSetup = $derived(
    neverSubscribed ||
      ((onManagedPlan || onEnterprisePlan) &&
        needsPaymentSetup(paymentIssues ?? [])),
  );

  async function handleManageCards() {
    const setup = paymentIssues?.length
      ? needsPaymentSetup(paymentIssues)
      : false;
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href, setup),
      "_blank",
    );
  }
</script>

<section>
  <h2 class="section-header">Payment methods</h2>
  <div class="section-card">
    <div class="card-content">
      {#if paymentIssues?.length}
        <div class="flex items-center gap-x-2">
          <CancelCircle className="text-red-600" size="14px" />
          <span class="text-sm text-red-600">
            {getPaymentIssueErrorText(paymentIssues)}
          </span>
        </div>
      {:else if hasPaymentCustomer}
        <!-- TODO: Show card details (brand, last4, expiry) from Stripe API -->
        <div class="flex items-center gap-x-3">
          <div class="card-icon">
            <svg
              class="w-5 h-5 text-fg-secondary"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                d="M2.5 4A1.5 1.5 0 001 5.5v1h18v-1A1.5 1.5 0 0017.5 4h-15zM19 8.5H1v6A1.5 1.5 0 002.5 16h15a1.5 1.5 0 001.5-1.5v-6zM3 12a1 1 0 011-1h3a1 1 0 110 2H4a1 1 0 01-1-1z"
              />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-fg-primary">
              Payment method on file
            </p>
            <p class="text-xs text-fg-tertiary">Manage your cards via Stripe</p>
          </div>
        </div>
      {:else}
        <span class="text-sm text-fg-tertiary">No payment method on file.</span>
      {/if}
    </div>
    {#if hasPaymentCustomer}
      <button class="manage-btn" onclick={handleManageCards}>
        Manage in Stripe
        <svg
          class="w-3.5 h-3.5"
          viewBox="0 0 12 12"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M2.5 9.5l7-7M4 2.5h6v6" />
        </svg>
      </button>
    {/if}
  </div>
</section>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }

  .section-card {
    @apply flex items-center justify-between border rounded-lg p-4 bg-surface-background;
    box-shadow:
      0px 1px 2px 0px rgba(0, 0, 0, 0.06),
      0px 1px 3px 0px rgba(0, 0, 0, 0.1);
  }

  .card-content {
    @apply flex items-center;
  }

  .card-icon {
    @apply flex items-center justify-center w-8 h-8 bg-surface-subtle rounded;
  }

  .manage-btn {
    @apply flex items-center gap-1.5 text-sm font-medium text-primary-600 border border-primary-500 rounded-sm px-4 py-2 bg-transparent cursor-pointer;
  }

  .manage-btn:hover {
    @apply bg-primary-50;
  }
</style>
