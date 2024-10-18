<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client";
  import { getPaymentIssueErrorText } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    fetchPaymentsPortalURL,
    fetchTeamPlan,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { showWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  /**
   * Landing page to upgrade a user to team plan.
   * Is set as a return url on stripe portal.
   */
  $: organization = $page.params.organization;
  let upgrading = false;
  let isRenew = false;
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: if (!$categorisedIssues.isLoading && !upgrading) {
    upgrade();
  }

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();

  async function upgrade() {
    // if there are still payment issues then do not upgrade
    if ($categorisedIssues.data?.payment?.length) {
      eventBus.emit("notification", {
        type: "error",
        message: `Please fix payment issues: ${getPaymentIssueErrorText($categorisedIssues.data.payment)}`,
        link: {
          text: "Update payment",
          href: await fetchPaymentsPortalURL(
            organization,
            getBillingUpgradeUrl($page, organization),
          ),
        },
      });
      return goto(`/${organization}`);
    }
    isRenew = !!$categorisedIssues.data?.cancelled;
    const teamPlan = await fetchTeamPlan();
    try {
      if (isRenew) {
        await $planRenewer.mutateAsync({
          organization,
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
          organization,
          data: {
            planName: teamPlan.name,
          },
        });
        showWelcomeToRillDialog.set(true);
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
    {#if !$categorisedIssues.isLoading}
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
