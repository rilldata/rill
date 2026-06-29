<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    getPaymentIssueErrorText,
    needsPaymentSetup,
  } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import {
    fetchPaymentsPortalURL,
    maybeFetchPublicPlanByName,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import {
    SELF_SERVE_PLANS,
    SELF_SERVE_PLANS_BY_NAME,
  } from "@rilldata/web-admin/features/billing/plans/plan-details";
  import { triggerWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let data: PageData;
  $: ({ cancelled, paymentIssues } = data);
  $: redirect = $page.url.searchParams.get("redirect");
  // The chosen plan is carried through the Stripe return URL (see getBillingUpgradeUrl).
  // Fall back to the first self-serve plan for older links that predate the plan chooser.
  $: planName = $page.url.searchParams.get("plan") ?? SELF_SERVE_PLANS[0].name;
  $: planDetails = SELF_SERVE_PLANS_BY_NAME[planName];
  $: planDisplayName = planDetails?.displayName ?? planName;

  /**
   * Landing page to upgrade a user to team plan.
   * Is set as a return url on stripe portal.
   */
  $: organization = $page.params.organization;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();

  async function upgrade() {
    // if there are still payment issues then do not upgrade
    if (paymentIssues.length) {
      eventBus.emit("notification", {
        type: "error",
        message: m.billing_fix_payment_issues({ details: getPaymentIssueErrorText(paymentIssues) }),
        link: {
          text: m.billing_update_payment(),
          href: await fetchPaymentsPortalURL(
            organization,
            getBillingUpgradeUrl($page, organization),
            needsPaymentSetup(paymentIssues),
          ),
        },
        options: {
          persisted: true,
        },
      });
      return goto(`/${organization}/-/settings/billing`);
    }
    const paidPlan = await maybeFetchPublicPlanByName(planName);
    if (!paidPlan) return goto(`/${organization}/-/settings/billing`);
    try {
      if (cancelled) {
        await $planRenewer.mutateAsync({
          org: organization,
          data: { planName },
        });
        eventBus.emit("notification", {
          type: "success",
          message: m.billing_plan_renewed({ plan: paidPlan.displayName }),
        });
      } else {
        await $planUpdater.mutateAsync({
          org: organization,
          data: { planName },
        });
        // if redirect is set then this page won't be active.
        // so this will lead to pop-in of the modal before navigating away
        if (!redirect) {
          triggerWelcomeToRillDialog(planName);
        }
      }
      void invalidateBillingInfo(organization);
    } catch {
      // TODO
    }
    if (redirect) {
      // redirect param could be on a different domain like the rill developer instance
      // so using goto won't work
      window.open(redirect, "_self");
    } else {
      return goto(`/${organization}`);
    }
  }

  onMount(() => upgrade());
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      {#if cancelled}
        {m.billing_renewing_plan({ plan: planDisplayName })}
      {:else}
        {m.billing_upgrading_plan({ plan: planDisplayName })}
      {/if}
    </CtaHeader>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
