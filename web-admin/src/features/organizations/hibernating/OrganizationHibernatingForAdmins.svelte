<script lang="ts">
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import StartTeamPlanDialog, {
    type TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { wakeAllProjects } from "@rilldata/web-admin/features/organizations/hibernating/wakeAllProjects";
  import { Button } from "@rilldata/web-common/components/button";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import AlertCircleIcon from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import MoonCircleOutline from "@rilldata/web-common/components/icons/MoonCircleOutline.svelte";

  export let organization: string;

  $: billingIssueMessage = useBillingIssueMessage(organization);
  let showStartTeamPlanDialog = false;
  let startTeamPlanType: TeamPlanDialogTypes = "base";
  let teamPlanEndDate = "";
  async function handleBannerCTAClick(issueMessage: BillingIssueMessage) {
    if (!issueMessage.cta) return;
    switch (issueMessage.cta.type) {
      case "upgrade":
        showStartTeamPlanDialog = true;
        startTeamPlanType = issueMessage.cta.teamPlanDialogType;
        teamPlanEndDate = issueMessage.cta.teamPlanEndDate ?? "";
        break;

      case "payment":
        window.open(
          await fetchPaymentsPortalURL(organization, window.location.href),
          "_self",
        );
        break;

      case "contact":
        window.Pylon("show");
        break;

      case "wake-projects":
        void wakeAllProjects(organization);
        break;
    }
  }

  const IconMap = {
    alert: AlertCircleIcon,
    check: CheckCircleOutline,
    sleep: MoonCircleOutline,
    loading: LoadingCircleOutline,
  };
</script>

{#if $billingIssueMessage.data}
  <div class="flex flex-col justify-center items-center gap-y-6 my-20">
    <div class="flex flex-col gap-y-2">
      {#if $billingIssueMessage.data.iconType in IconMap}
        <!-- TODO: gradient -->
        <svelte:component
          this={IconMap[$billingIssueMessage.data.iconType]}
          size="104px"
          className="text-slate-300"
        />
      {/if}
      <CTAHeader variant="bold">{$billingIssueMessage.data.title}</CTAHeader>
      <CTAMessage>
        {$billingIssueMessage.data.description}
      </CTAMessage>
    </div>
    {#if $billingIssueMessage.data.cta}
      <Button
        type="secondary"
        wide
        on:click={() => handleBannerCTAClick($billingIssueMessage.data)}
      >
        {$billingIssueMessage.data.cta.text}
      </Button>
    {/if}
    <CTANeedHelp />
  </div>
{/if}

<StartTeamPlanDialog
  bind:open={showStartTeamPlanDialog}
  type={startTeamPlanType}
  endDate={teamPlanEndDate}
  {organization}
/>
