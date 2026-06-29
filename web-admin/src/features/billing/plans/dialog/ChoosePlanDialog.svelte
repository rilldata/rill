<script lang="ts">
  import { page } from "$app/state";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types.ts";
  import { getSubscriptionResumedText } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import {
    resolvePlanHighlights,
    SELF_SERVE_PLANS,
    getTranslatedPlanDisplayName,
    getTranslatedPlanTagline,
    getTranslatedPlanPriceUnit,
  } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { upgradeToPlan } from "@rilldata/web-admin/features/billing/plans/upgrade-to-plan.ts";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors.ts";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListPublicBillingPlans,
  } from "@rilldata/web-admin/client";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    open = $bindable(false),
    type,
    organization,
    endDate = "",
  }: {
    open: boolean;
    type: TeamPlanDialogTypes;
    organization: string;
    endDate?: string;
  } = $props();

  let title = $derived.by(() => {
    switch (type) {
      case "size":
        return m.billing_dialog_title_size();
      case "org":
        return m.billing_dialog_title_org();
      case "proj":
        return m.billing_dialog_title_proj();
      case "renew":
        return m.billing_dialog_title_renew();
      case "trial-expired":
        return m.billing_dialog_title_trial_expired();
      case "base":
      default:
        return m.billing_choose_a_plan();
    }
  });

  let description = $derived(
    type === "renew"
      ? m.billing_cycle_will_resume({ resumeText: getSubscriptionResumedText(endDate) })
      : m.billing_choosing_plan_ends_trial(),
  );

  let orgQuery = $derived(createAdminServiceGetOrganization(organization));
  let currentPlanName = $derived($orgQuery.data?.organization?.billingPlanName);
  let currentPlanQuota = $derived($orgQuery.data?.organization?.quotas);

  let categorisedIssuesQuery = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let categorisedIssues = $derived($categorisedIssuesQuery.data);
  let redirect = $derived(page.url.searchParams.get("redirect"));

  const plansQuery = $derived(
    createAdminServiceListPublicBillingPlans({
      query: {
        enabled: open,
      },
    }),
  );
  let plans = $derived($plansQuery.data?.plans ?? []);

  let loadingPlan = $state<string | null>(null);
  let fetchError = $state<string | null>(null);

  async function handleUpgradePlan(planName: string) {
    if (!categorisedIssues) return;
    loadingPlan = planName;
    fetchError = null;
    try {
      await upgradeToPlan(organization, planName, categorisedIssues, redirect);
      // Only close if the upgrade was successful.
      open = false;
    } catch (e) {
      fetchError = extractErrorMessage(e);
    }
    loadingPlan = null;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent class="max-w-3xl">
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <p class="text-sm text-fg-tertiary">{description}</p>
      {#if fetchError}
        <div class="text-red-500 text-sm py-px">{fetchError}</div>
      {/if}
    </AlertDialogHeader>

    <div class="h-[400px] w-full">
      {#if $plansQuery.isPending}
        <div class="flex size-full items-center justify-center">
          <Spinner status={EntityStatus.Running} size="2rem" duration={725} />
        </div>
      {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-2">
          {#each SELF_SERVE_PLANS as plan (plan.name)}
            {@const isCurrentPlan = plan.name === currentPlanName}
            {@const highlights = resolvePlanHighlights(
              plan,
              isCurrentPlan && currentPlanQuota
                ? currentPlanQuota
                : (plans.find((p) => p.name === plan.name)?.quotas ?? {}),
            )}

            <div
              class="flex flex-col border rounded-xl p-5 gap-3"
              class:border-primary-500={plan.recommended}
            >
              <div class="flex items-center justify-between">
                <span class="text-lg font-semibold text-fg-primary">
                  {getTranslatedPlanDisplayName(plan.name)}
                </span>
                {#if plan.recommended}
                  <span
                    class="text-xs font-semibold text-primary-600 bg-primary-50 rounded-full px-2 py-0.5"
                  >
                    {m.billing_recommended()}
                  </span>
                {/if}
              </div>

              <div class="flex items-baseline gap-1">
                <span class="text-2xl font-semibold text-fg-primary">
                  {plan.price}
                </span>
                <span class="text-sm text-fg-tertiary">{getTranslatedPlanPriceUnit()}</span>
              </div>
              <p class="text-sm text-fg-tertiary">{getTranslatedPlanTagline(plan.name)}</p>

              <ul class="flex flex-col gap-1.5 mt-1 grow">
                {#each highlights as highlight (highlight)}
                  <li class="flex items-start gap-2 text-sm text-fg-secondary">
                    <span class="text-primary-600 mt-0.5">
                      <Check size="14px" />
                    </span>
                    {highlight}
                  </li>
                {/each}
              </ul>

              <Button
                type={plan.recommended ? "primary" : "secondary"}
                wide
                loading={loadingPlan === plan.name}
                disabled={loadingPlan !== null || isCurrentPlan}
                onClick={() => handleUpgradePlan(plan.name)}
              >
                {#if isCurrentPlan}{m.billing_current()}{:else}{m.billing_choose_plan_name({ planName: getTranslatedPlanDisplayName(plan.name) })}{/if}
              </Button>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <AlertDialogFooter class="mt-3">
      <Button type="text" onClick={() => (open = false)}>{m.billing_close()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
