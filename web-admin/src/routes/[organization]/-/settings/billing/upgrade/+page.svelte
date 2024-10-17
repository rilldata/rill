<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    fetchPaymentsPortalURL,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  /**
   * Landing page to upgrade a user to team plan.
   * If there are no billing issues calls upgrade/renew. Else redirects to payment portal.
   * Typically set as a CTA in emails.
   */

  $: organization = $page.params.organization;

  let upgrading = false;
  let isRenew = false;
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: if (!$categorisedIssues.isLoading && !upgrading) {
    upgrade();
  }

  async function upgrade() {
    if (!$categorisedIssues.data?.payment?.length) {
      return goto(`/${organization}/-/settings/billing/upgrade/callback`);
    }

    window.open(
      await fetchPaymentsPortalURL(
        organization,
        getBillingUpgradeUrl($page, organization),
      ),
      "_self",
    );
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
