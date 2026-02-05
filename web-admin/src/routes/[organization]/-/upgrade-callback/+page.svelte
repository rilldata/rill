<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import {
    fetchPaymentsPortalURL,
    fetchTeamPlan,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { showWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ cancelled, paymentIssues } = data);
  $: redirect = $page.url.searchParams.get("redirect");

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
        message: `Please fix payment issues: ${getPaymentIssueErrorText(paymentIssues)}`,
        link: {
          text: "Update payment",
          href: await fetchPaymentsPortalURL(
            organization,
            getBillingUpgradeUrl($page, organization),
          ),
        },
        options: {
          persisted: true,
        },
      });
      return goto(`/${organization}/-/settings/billing`);
    }
    const teamPlan = await fetchTeamPlan();
    try {
      if (cancelled) {
        await $planRenewer.mutateAsync({
          org: organization,
          data: {
            planName: teamPlan.name,
          },
        });
        eventBus.emit("notification", {
          type: "success",
          message: "Your Team plan was renewed",
        });
      } else {
        await $planUpdater.mutateAsync({
          org: organization,
          data: {
            planName: teamPlan.name,
          },
        });
        // if redirect is set then this page won't be active.
        // so this will lead to pop-in of the modal before navigating away
        if (!redirect) {
          showWelcomeToRillDialog.set(true);
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
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader variant="bold">
      {#if cancelled}
        Renewing team plan...
      {:else}
        Upgrading to team plan...
      {/if}
    </CtaHeader>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
