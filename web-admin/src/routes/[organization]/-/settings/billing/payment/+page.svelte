<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { V1BillingIssueType } from "@rilldata/web-admin/client";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import PlusIcon from "@rilldata/web-common/components/icons/PlusIcon.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ organization, teamPlan, issues } = data);
  $: returnUrl = `${$page.url.protocol}//${$page.url.host}/${organization}/-/upgrade-callback`;

  // Parse billing issues to determine what's missing
  $: hasNoPaymentMethod = issues?.some(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD,
  );
  $: hasNoBillableAddress = issues?.some(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS,
  );
  $: hasPaymentFailed = issues?.some(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED,
  );

  // Determine what needs to be done
  $: needsPaymentMethod = hasNoPaymentMethod;
  $: needsBillingAddress = hasNoBillableAddress;
  $: needsPaymentFix = hasPaymentFailed;
  $: allComplete =
    !needsPaymentMethod && !needsBillingAddress && !needsPaymentFix;

  let loading = false;

  async function handleContinueToPayment() {
    loading = true;
    try {
      const portalUrl = await fetchPaymentsPortalURL(organization, returnUrl);
      window.open(portalUrl, "_self");
    } catch {
      loading = false;
    }
  }

  function handleGoBack() {
    goto(`/${organization}/-/settings/billing`);
  }
</script>

<div class="flex flex-col md:flex-row min-h-[600px] rounded-lg overflow-hidden border">
  <!-- Left Side: Plan Details -->
  <div class="flex flex-col p-8 bg-primary-800 text-white md:w-1/2">
    <div class="flex flex-col gap-4 mb-6">
      <div class="text-primary-400">
        <svg
          width="32"
          height="32"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M4 4h4v16H4V4zm6 0h4v16h-4V4zm6 0h4v16h-4V4z"
            fill="currentColor"
          />
        </svg>
      </div>
      <h1 class="text-2xl font-semibold">
        Subscribe to {teamPlan?.displayName || "Team"}
      </h1>
    </div>

    <div class="mb-6">
      <div class="flex items-baseline gap-2">
        <span class="text-5xl font-bold">$250</span>
        <span class="text-sm text-primary-300 leading-tight">per<br />month</span
        >
      </div>
    </div>

    <div class="bg-primary-700/50 rounded-lg p-4 mb-6">
      <div class="flex justify-between items-center mb-1">
        <div class="font-medium">{teamPlan?.displayName || "Team"}</div>
        <div class="font-semibold">$250.00</div>
      </div>
      <div class="text-sm text-primary-300">Rill Team Monthly subscription</div>
      <div class="text-xs text-primary-400 mt-1">Billed monthly</div>
    </div>

    <div class="mb-6">
      <div class="text-sm text-primary-300 mb-2">Includes:</div>
      <ul class="space-y-2">
        <li class="flex items-center gap-2 text-sm">
          <span class="text-green-400 flex-shrink-0"><Check size="16px" /></span
          >
          <span>10 GB of data included</span>
        </li>
        <li class="flex items-center gap-2 text-sm">
          <span class="text-green-400 flex-shrink-0"><Check size="16px" /></span
          >
          <span>$25/GB for additional data</span>
        </li>
        <li class="flex items-center gap-2 text-sm">
          <span class="text-green-400 flex-shrink-0"><Check size="16px" /></span
          >
          <span>Unlimited projects (max 50 GB each)</span>
        </li>
        <li class="flex items-center gap-2 text-sm">
          <span class="text-green-400 flex-shrink-0"><Check size="16px" /></span
          >
          <span>Unlimited users</span>
        </li>
      </ul>
    </div>

    <div class="border-t border-primary-600 pt-4 mt-auto">
      <div class="flex justify-between items-center py-1">
        <span class="text-sm text-primary-300">Subtotal</span>
        <span class="font-medium">$250.00</span>
      </div>
      <div class="flex justify-between items-center py-1">
        <span class="text-sm text-primary-300 flex items-center gap-1">
          Tax
          <InfoCircle size="14px" />
        </span>
        <span class="text-sm text-primary-400">Enter address to calculate</span>
      </div>
      <div class="border-t border-primary-600 my-2"></div>
      <div class="flex justify-between items-center py-1">
        <span class="font-medium">Total due today</span>
        <span class="text-xl font-bold">$250.00</span>
      </div>
    </div>

    <div class="mt-4 text-sm">
      <a
        href="https://www.rilldata.com/pricing"
        target="_blank"
        rel="noreferrer noopener"
        class="text-primary-300 hover:text-primary-200 underline"
      >
        See full pricing details
      </a>
    </div>
  </div>

  <!-- Right Side: Payment Requirements -->
  <div class="flex flex-col p-8 bg-surface-background md:w-1/2">
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-fg-primary mb-2">
        Complete your subscription
      </h2>
      <p class="text-sm text-fg-secondary">
        To start your Team plan, please complete the following in the Stripe
        billing portal:
      </p>
    </div>

    <div class="space-y-4 mb-6">
      <!-- Payment Method Requirement -->
      <div
        class="flex items-start gap-3 p-4 rounded-lg border {!needsPaymentMethod
          ? 'border-green-200 bg-green-50'
          : 'bg-surface-subtle'}"
      >
        <div
          class="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center {!needsPaymentMethod
            ? 'bg-green-100 text-green-600'
            : 'bg-surface-background text-fg-tertiary'}"
        >
          {#if needsPaymentMethod}
            <PlusIcon size="20px" />
          {:else}
            <Check size="20px" />
          {/if}
        </div>
        <div class="flex-1">
          <div class="font-medium text-fg-primary">Payment method</div>
          <div class="text-sm text-fg-secondary">
            {#if needsPaymentMethod}
              Add a credit card or other payment method
            {:else}
              Payment method added
            {/if}
          </div>
        </div>
        {#if needsPaymentMethod}
          <div
            class="text-xs px-2 py-1 rounded-full font-medium bg-yellow-100 text-yellow-700"
          >
            Required
          </div>
        {:else}
          <div
            class="text-xs px-2 py-1 rounded-full font-medium bg-green-100 text-green-700"
          >
            Complete
          </div>
        {/if}
      </div>

      <!-- Billing Address Requirement -->
      <div
        class="flex items-start gap-3 p-4 rounded-lg border {!needsBillingAddress
          ? 'border-green-200 bg-green-50'
          : 'bg-surface-subtle'}"
      >
        <div
          class="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center {!needsBillingAddress
            ? 'bg-green-100 text-green-600'
            : 'bg-surface-background text-fg-tertiary'}"
        >
          {#if needsBillingAddress}
            <InfoCircle size="20px" />
          {:else}
            <Check size="20px" />
          {/if}
        </div>
        <div class="flex-1">
          <div class="font-medium text-fg-primary">Billing information</div>
          <div class="text-sm text-fg-secondary">
            {#if needsBillingAddress}
              Add your billing address for tax calculation
            {:else}
              Billing information added
            {/if}
          </div>
        </div>
        {#if needsBillingAddress}
          <div
            class="text-xs px-2 py-1 rounded-full font-medium bg-yellow-100 text-yellow-700"
          >
            Required
          </div>
        {:else}
          <div
            class="text-xs px-2 py-1 rounded-full font-medium bg-green-100 text-green-700"
          >
            Complete
          </div>
        {/if}
      </div>

      <!-- Payment Failed Error -->
      {#if needsPaymentFix}
        <div
          class="flex items-start gap-3 p-4 rounded-lg border border-red-200 bg-red-50"
        >
          <div
            class="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center bg-red-100 text-red-600"
          >
            <CancelCircle size="20px" />
          </div>
          <div class="flex-1">
            <div class="font-medium text-fg-primary">Payment issue</div>
            <div class="text-sm text-fg-secondary">
              A previous payment failed. Please update your payment method.
            </div>
          </div>
          <div
            class="text-xs px-2 py-1 rounded-full font-medium bg-red-100 text-red-700"
          >
            Action required
          </div>
        </div>
      {/if}
    </div>

    {#if allComplete}
      <div
        class="flex items-center gap-2 p-4 rounded-lg bg-green-50 text-green-700 text-sm mb-6"
      >
        <Check size="20px" />
        <span
          >All requirements are complete. You can proceed to subscribe.</span
        >
      </div>
    {/if}

    <div class="mb-6">
      <div class="text-sm text-fg-tertiary mb-2">Accepted payment methods</div>
      <div class="flex gap-2">
        <div
          class="text-xs px-2 py-1 rounded border bg-surface-background text-fg-secondary"
        >
          Visa
        </div>
        <div
          class="text-xs px-2 py-1 rounded border bg-surface-background text-fg-secondary"
        >
          Mastercard
        </div>
        <div
          class="text-xs px-2 py-1 rounded border bg-surface-background text-fg-secondary"
        >
          Amex
        </div>
        <div
          class="text-xs px-2 py-1 rounded border bg-surface-background text-fg-secondary"
        >
          Discover
        </div>
      </div>
    </div>

    <div class="flex gap-3 mt-auto">
      <Button type="secondary" onClick={handleGoBack}>Back to billing</Button>
      <Button type="primary" onClick={handleContinueToPayment} {loading}>
        Continue to payment
      </Button>
    </div>

    <div class="flex items-center gap-2 mt-6 text-xs text-fg-tertiary">
      <span>Powered by</span>
      <span class="font-semibold text-fg-secondary">Stripe</span>
      <span class="text-fg-disabled">|</span>
      <a
        href="https://stripe.com/legal"
        target="_blank"
        rel="noreferrer"
        class="hover:text-fg-secondary">Terms</a
      >
      <a
        href="https://stripe.com/privacy"
        target="_blank"
        rel="noreferrer"
        class="hover:text-fg-secondary">Privacy</a
      >
    </div>
  </div>
</div>
