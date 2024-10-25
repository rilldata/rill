<script lang="ts">
  import { BillingCTAHandler } from "@rilldata/web-admin/features/billing/BillingCTAHandler";
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
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
  $: billingCTAHandler = new BillingCTAHandler(organization);
  $: ({
    showStartTeamPlanDialog,
    startTeamPlanType,
    teamPlanEndDate,
    wakingProjects,
  } = billingCTAHandler);
  let issueForHibernation: BillingIssueMessage;
  $: if ($billingIssueMessage.data) {
    if ($billingIssueMessage.data.type === "error") {
      // only show the hibernating message if it is an error
      issueForHibernation = $billingIssueMessage.data;
    } else {
      // else who a CTA for waking up projects
      issueForHibernation = {
        type: "default",
        iconType: "sleep",
        title: "Your projects are hibernating",
        description: "",
        cta: {
          text: "Wake projects",
          type: "wake-projects",
        },
      };
    }
  }

  const IconMap = {
    alert: AlertCircleIcon,
    check: CheckCircleOutline,
    sleep: MoonCircleOutline,
    loading: LoadingCircleOutline,
  };
</script>

{#if $wakingProjects}
  <div class="flex flex-col justify-center items-center gap-y-6 my-20">
    <div class="flex flex-col gap-y-2">
      <LoadingCircleOutline size="104px" className="text-slate-300" />
      <CTAHeader>Hang tight! We're waking up your projects...</CTAHeader>
      <CTANeedHelp />
    </div>
  </div>
{:else if issueForHibernation}
  <div class="flex flex-col justify-center items-center gap-y-6 my-20">
    <div class="flex flex-col gap-y-2">
      {#if issueForHibernation.iconType in IconMap}
        <!-- TODO: gradient -->
        <svelte:component
          this={IconMap[issueForHibernation.iconType]}
          size="104px"
          className="text-slate-300"
          gradientStopColor="slate-200"
        />
      {/if}
      <CTAHeader variant="bold">
        {issueForHibernation.title.replace(/\.$/, "")}
      </CTAHeader>
      {#if issueForHibernation.description}
        <CTAMessage>
          {issueForHibernation.description.replace(/\.$/, "")}
        </CTAMessage>
      {/if}
    </div>
    {#if issueForHibernation.cta}
      <Button
        type="secondary"
        wide
        on:click={() => billingCTAHandler.handle(issueForHibernation)}
      >
        {issueForHibernation.cta.text}
      </Button>
    {/if}
    <CTANeedHelp />
  </div>
{/if}

<StartTeamPlanDialog
  bind:open={$showStartTeamPlanDialog}
  type={$startTeamPlanType}
  endDate={$teamPlanEndDate}
  {organization}
/>
