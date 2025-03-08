<script lang="ts">
  import { BillingCTAHandler } from "@rilldata/web-admin/features/billing/BillingCTAHandler";
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { BannerSlot } from "@rilldata/web-common/lib/event-bus/events";

  export let organization: string;

  $: billingIssueMessage = useBillingIssueMessage(organization);
  $: billingCTAHandler = new BillingCTAHandler(organization);
  $: ({ showStartTeamPlanDialog, startTeamPlanType, teamPlanEndDate } =
    billingCTAHandler);

  function showBillingIssueBanner(message: BillingIssueMessage | undefined) {
    if (!message) {
      eventBus.emit("banner", {
        slot: BannerSlot.Billing,
        message: null,
      });
      return;
    }

    eventBus.emit("banner", {
      slot: BannerSlot.Billing,
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
