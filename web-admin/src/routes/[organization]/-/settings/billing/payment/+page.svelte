<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    createPaymentCheckoutSessionURL,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";

  $: organization = $page.params.organization;
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment ?? [];
  $: hasNoPaymentMethod = paymentIssues.some(
    (i) => i.type === "BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD",
  );
  $: hasNoBillingAddress = paymentIssues.some(
    (i) => i.type === "BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS",
  );
  $: hasPaymentFailed = paymentIssues.some(
    (i) => i.type === "BILLING_ISSUE_TYPE_PAYMENT_FAILED",
  );

  let loading = false;

  async function handleContinueToCheckout() {
    loading = true;
    try {
      const successUrl = getBillingUpgradeUrl($page, organization);
      const cancelUrl = `${$page.url.protocol}//${$page.url.host}/${organization}/-/settings/billing/payment`;
      const checkoutUrl = await createPaymentCheckoutSessionURL(
        organization,
        successUrl,
        cancelUrl,
      );
      if (checkoutUrl) {
        window.open(checkoutUrl, "_self");
        return;
      }
      loading = false;
      eventBus.emit("notification", {
        type: "error",
        message: "Failed to open payment page. Please try again.",
      });
    } catch (e) {
      console.error("Failed to open checkout:", e);
      loading = false;
      eventBus.emit("notification", {
        type: "error",
        message: "Failed to open payment page. Please try again.",
      });
    }
  }

  function handleBack() {
    goto(`/${organization}/-/settings/billing`);
  }
</script>

<div class="flex min-h-[600px] w-full">
  <!-- Left Panel - Dark theme with pricing -->
  <div class="w-1/2 bg-gray-900 text-white p-8 flex flex-col">
    <div class="mb-8">
      <button
        on:click={handleBack}
        class="flex items-center gap-1 text-gray-400 hover:text-white transition-colors text-sm"
      >
        ← Back to billing
      </button>
    </div>

    <div class="flex-1">
      <p class="text-gray-400 text-sm mb-2">Subscribe to</p>
      <h1 class="text-2xl font-semibold mb-1">Team Plan</h1>
      <div class="flex items-baseline gap-2 mb-8">
        <span class="text-4xl font-bold">$250</span>
        <span class="text-gray-400">per month</span>
      </div>

      <div class="bg-gray-800 rounded-lg p-4 mb-6">
        <div class="flex justify-between items-center mb-3">
          <div>
            <p class="font-medium">Team Plan</p>
            <p class="text-gray-400 text-sm">Monthly subscription</p>
          </div>
          <span class="font-semibold">$250.00</span>
        </div>
        <div class="border-t border-gray-700 pt-3 space-y-2 text-sm text-gray-300">
          <div class="flex justify-between">
            <span>10 GB data included</span>
          </div>
          <div class="flex justify-between">
            <span>$25/GB for additional data</span>
          </div>
          <div class="flex justify-between">
            <span>Unlimited projects (50 GB each)</span>
          </div>
          <div class="flex justify-between">
            <span>Unlimited users</span>
          </div>
        </div>
      </div>

      <div class="border-t border-gray-700 pt-4">
        <div class="flex justify-between text-lg font-semibold">
          <span>Starting at</span>
          <span>$250.00/month</span>
        </div>
        <p class="text-gray-400 text-sm mt-1">
          Usage-based billing. See <a
            href="https://www.rilldata.com/pricing"
            target="_blank"
            rel="noreferrer noopener"
            class="text-primary-400 hover:underline">pricing details</a
          >.
        </p>
      </div>
    </div>
  </div>

  <!-- Right Panel - Light theme with requirements -->
  <div class="w-1/2 bg-white p-8 flex flex-col">
    <div class="flex-1">
      <h2 class="text-lg font-semibold mb-6">Complete your setup</h2>

      <div class="space-y-4">
        <!-- Payment Method Status -->
        <div
          class="flex items-start gap-3 p-4 rounded-lg border {hasNoPaymentMethod
            ? 'border-red-200 bg-red-50'
            : 'border-green-200 bg-green-50'}"
        >
          <div class="mt-0.5">
            {#if hasNoPaymentMethod}
              <CancelCircle className="text-red-500" size="20px" />
            {:else}
              <Check className="text-green-600" size="20px" />
            {/if}
          </div>
          <div>
            <p class="font-medium {hasNoPaymentMethod ? 'text-red-800' : 'text-green-800'}">
              Payment method
            </p>
            <p class="text-sm {hasNoPaymentMethod ? 'text-red-600' : 'text-green-600'}">
              {hasNoPaymentMethod ? "Required - Add a payment method" : "Payment method added"}
            </p>
          </div>
        </div>

        <!-- Billing Address Status -->
        <div
          class="flex items-start gap-3 p-4 rounded-lg border {hasNoBillingAddress
            ? 'border-red-200 bg-red-50'
            : 'border-green-200 bg-green-50'}"
        >
          <div class="mt-0.5">
            {#if hasNoBillingAddress}
              <CancelCircle className="text-red-500" size="20px" />
            {:else}
              <Check className="text-green-600" size="20px" />
            {/if}
          </div>
          <div>
            <p class="font-medium {hasNoBillingAddress ? 'text-red-800' : 'text-green-800'}">
              Billing address
            </p>
            <p class="text-sm {hasNoBillingAddress ? 'text-red-600' : 'text-green-600'}">
              {hasNoBillingAddress ? "Required - Add billing address" : "Billing address added"}
            </p>
          </div>
        </div>

        <!-- Payment Failed Status (if applicable) -->
        {#if hasPaymentFailed}
          <div class="flex items-start gap-3 p-4 rounded-lg border border-red-200 bg-red-50">
            <div class="mt-0.5">
              <CancelCircle className="text-red-500" size="20px" />
            </div>
            <div>
              <p class="font-medium text-red-800">Payment issue</p>
              <p class="text-sm text-red-600">
                A previous payment failed. Please update your payment method.
              </p>
            </div>
          </div>
        {/if}
      </div>

      <!-- Payment method icons -->
      <div class="mt-8">
        <p class="text-sm text-gray-500 mb-3">Accepted payment methods</p>
        <div class="flex gap-2">
          <div class="px-3 py-1.5 bg-gray-100 rounded text-xs font-medium text-gray-700">
            Visa
          </div>
          <div class="px-3 py-1.5 bg-gray-100 rounded text-xs font-medium text-gray-700">
            Mastercard
          </div>
          <div class="px-3 py-1.5 bg-gray-100 rounded text-xs font-medium text-gray-700">
            Amex
          </div>
          <div class="px-3 py-1.5 bg-gray-100 rounded text-xs font-medium text-gray-700">
            Discover
          </div>
        </div>
      </div>
    </div>

    <!-- CTA Button -->
    <div class="mt-8">
      <Button
        type="primary"
        onClick={handleContinueToCheckout}
        {loading}
        wide
      >
        Continue to payment
      </Button>
      <p class="text-xs text-gray-500 mt-3 text-center">
        You'll be redirected to Stripe to securely add your payment details.
      </p>
    </div>
  </div>
</div>
