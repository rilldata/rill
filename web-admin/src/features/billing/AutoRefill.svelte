<script lang="ts">
  import {
    createAdminServiceGetAutoRefillSettings,
    createAdminServiceGetBillingSubscription,
    createAdminServiceUpdateAutoRefillSettings,
  } from "@rilldata/web-admin/client";
  import { isFreePlan, isTrialPlan } from "./plans/utils";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: plan = $subscriptionQuery?.data?.subscription?.plan;
  $: isFree = plan && isFreePlan(plan.name);
  $: isTrial = plan && isTrialPlan(plan.name);

  $: settingsQuery = createAdminServiceGetAutoRefillSettings(organization);

  // Local form state
  let enabled: boolean;
  let threshold: string;
  let amount: string;
  let initialized = false;

  // Sync from server once loaded
  $: if ($settingsQuery?.data && !initialized) {
    enabled = $settingsQuery.data.enabled ?? false;
    threshold = $settingsQuery.data.threshold
      ? String($settingsQuery.data.threshold)
      : "50";
    amount = $settingsQuery.data.amount
      ? String($settingsQuery.data.amount)
      : "100";
    initialized = true;
  }

  $: dirty =
    initialized &&
    $settingsQuery?.data &&
    (enabled !== ($settingsQuery.data.enabled ?? false) ||
      (enabled &&
        (parseFloat(threshold) !== ($settingsQuery.data.threshold ?? 0) ||
          parseFloat(amount) !== ($settingsQuery.data.amount ?? 0))));

  const mutation = createAdminServiceUpdateAutoRefillSettings();

  let saving = false;
  let error = "";

  async function handleSave() {
    error = "";
    const t = parseFloat(threshold);
    const a = parseFloat(amount);
    if (enabled && (isNaN(t) || t <= 0)) {
      error = "Threshold must be a positive number.";
      return;
    }
    if (enabled && (isNaN(a) || a <= 0)) {
      error = "Amount must be a positive number.";
      return;
    }

    saving = true;
    try {
      await $mutation.mutateAsync({
        org: organization,
        data: {
          enabled,
          threshold: enabled ? t : 0,
          amount: enabled ? a : 0,
        },
      });
      // Reset initialized to re-sync from server
      initialized = false;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : "Failed to save settings";
    } finally {
      saving = false;
    }
  }
</script>

{#if !$subscriptionQuery?.isLoading && !isFree && !isTrial}
  <SettingsContainer title="Auto-Refill Credits">
    <div slot="body" class="auto-refill-body">
      <p class="description">
        Automatically top up your credit balance when it drops below a
        threshold. A charge will be made to your payment method on file.
      </p>

      <label class="toggle-row">
        <input type="checkbox" bind:checked={enabled} />
        <span>Enable auto-refill</span>
      </label>

      {#if enabled}
        <div class="fields">
          <label class="field">
            <span class="field-label">When balance drops below</span>
            <div class="input-wrapper">
              <span class="input-prefix">$</span>
              <input
                type="number"
                min="1"
                step="any"
                bind:value={threshold}
                class="field-input"
                placeholder="50"
              />
            </div>
          </label>

          <label class="field">
            <span class="field-label">Top up amount</span>
            <div class="input-wrapper">
              <span class="input-prefix">$</span>
              <input
                type="number"
                min="1"
                step="any"
                bind:value={amount}
                class="field-input"
                placeholder="100"
              />
            </div>
          </label>
        </div>
      {/if}

      {#if error}
        <p class="error-text">{error}</p>
      {/if}
    </div>

    <Button
      slot="action"
      type="secondary"
      disabled={!dirty || saving}
      onClick={handleSave}
    >
      {saving ? "Saving..." : "Save"}
    </Button>
  </SettingsContainer>
{/if}

<style lang="postcss">
  .auto-refill-body {
    @apply flex flex-col gap-3;
  }
  .description {
    @apply text-sm text-fg-tertiary;
  }
  .toggle-row {
    @apply flex items-center gap-2 cursor-pointer text-sm text-fg-primary;
  }
  .toggle-row input[type="checkbox"] {
    @apply w-4 h-4 cursor-pointer;
  }
  .fields {
    @apply flex flex-col gap-3 pl-6;
  }
  .field {
    @apply flex flex-col gap-1;
  }
  .field-label {
    @apply text-xs text-fg-secondary font-medium;
  }
  .input-wrapper {
    @apply flex items-center border border-border rounded-md overflow-hidden w-40;
  }
  .input-prefix {
    @apply px-2 text-sm text-fg-tertiary bg-surface-subtle border-r border-border;
  }
  .field-input {
    @apply px-2 py-1.5 text-sm text-fg-primary bg-transparent border-none outline-none w-full;
  }
  .error-text {
    @apply text-xs text-red-600;
  }
</style>
