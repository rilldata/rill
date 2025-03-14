<script lang="ts">
  import { BillingCTAHandler } from "@rilldata/web-admin/features/billing/BillingCTAHandler";
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import {
    BillingBannerID,
    BillingBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;

  $: billingIssueMessage = useBillingIssueMessage(organization);
  $: billingCTAHandler = new BillingCTAHandler(organization);
  $: ({ showStartTeamPlanDialog, startTeamPlanType, teamPlanEndDate } =
    billingCTAHandler);

  function showBillingIssueBanner(message: BillingIssueMessage | undefined) {
    if (!message) {
      eventBus.emit("remove-banner", BillingBannerID);
      return;
    }

    eventBus.emit("add-banner", {
      id: BillingBannerID,
      priority: BillingBannerPriority,
      message: {
        type: message.type,
        message: message.title + " " + message.description,
        iconType: message.iconType,
        ...(message.cta
          ? {
              cta: {
                type: "button",
                text: message.cta.text + " ->",
                onClick() {
                  return billingCTAHandler.handle(message);
                },
              },
            }
          : {}),
      },
    });
  }

  $: showBillingIssueBanner($billingIssueMessage.data);
</script>

<StartTeamPlanDialog
  bind:open={$showStartTeamPlanDialog}
  type={$startTeamPlanType}
  endDate={$teamPlanEndDate}
  {organization}
/>
