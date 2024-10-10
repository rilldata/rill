<script lang="ts">
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import StartTeamPlanDialog, {
    type TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

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

      case "wake-projects":
        // TODO
        break;
    }
  }

  $: if ($billingIssueMessage.data) {
    eventBus.emit("banner", {
      type: $billingIssueMessage.data.type,
      message:
        $billingIssueMessage.data.title +
        " " +
        $billingIssueMessage.data.description,
      iconType: $billingIssueMessage.data.iconType,
      ...($billingIssueMessage.data.cta
        ? {
            cta: {
              type: "button",
              text: $billingIssueMessage.data.cta.text + "->",
              onClick() {
                void handleBannerCTAClick($billingIssueMessage.data);
              },
            },
          }
        : {}),
    });
  } else {
    // when switching orgs we need to make sure we clear previous org's banner.
    // TODO: could this interfere with other banners?
    eventBus.emit("banner", {
      type: "clear",
      message: "",
      iconType: "none",
    });
  }
</script>

<StartTeamPlanDialog
  bind:open={showStartTeamPlanDialog}
  type={startTeamPlanType}
  endDate={teamPlanEndDate}
  {organization}
/>
