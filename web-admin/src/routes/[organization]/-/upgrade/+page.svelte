<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceListOrganizationBillingIssues,
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client";
  import {
    getPaymentIssueErrorText,
    getPaymentIssues,
  } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { getCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue.js";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    fetchPaymentsPortalURL,
    fetchTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  $: organization = $page.params.organization;

  let upgrading = false;
  let isRenew = false;
  $: issuesQuery =
    createAdminServiceListOrganizationBillingIssues(organization);
  $: if (!$issuesQuery.isLoading && !upgrading) {
    upgrade();
  }

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();
  async function upgrade() {
    const paymentIssues = getPaymentIssues($issuesQuery.data?.issues ?? []);
    // if there are still payment issues then do not upgrade
    if (paymentIssues.length) {
      eventBus.emit("notification", {
        type: "error",
        message: `Please fix payment issues: ${getPaymentIssueErrorText(paymentIssues)}`,
        link: {
          text: "Update payment",
          href: await fetchPaymentsPortalURL(
            organization,
            window.location.href,
          ),
        },
        options: {
          persisted: true, // TODO: this is not honoured when link is added
        },
      });
      return goto(`/${organization}`);
    }

    isRenew = !!getCancelledIssue($issuesQuery.data?.issues ?? []);
    const teamPlan = await fetchTeamPlan();

    try {
      if (isRenew) {
        await $planRenewer.mutateAsync({
          organization,
          data: {
            planName: teamPlan.name,
          },
        });
        // TODO: show welcome to rill
      } else {
        await $planUpdater.mutateAsync({
          organization,
          data: {
            planName: teamPlan.name,
          },
        });
      }
      void invalidateBillingInfo(organization);
    } catch {
      // TODO
    }
    return goto(`/${organization}`);
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    {#if !$issuesQuery.isLoading}
      <CtaHeader variant="bold">
        {#if isRenew}
          Renewing team plan...
        {:else}
          Upgrading to team plan...
        {/if}
      </CtaHeader>
    {/if}
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
