<script lang="ts">
  import { getAdminServiceListProjectsForOrganizationQueryKey } from "@rilldata/web-admin/client";
  import { BannerCTAHandler } from "@rilldata/web-admin/features/billing/banner/BannerCTAHandler";
  import {
    type BillingIssueMessage,
    useBillingIssueMessage,
  } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;

  $: billingIssueMessage = useBillingIssueMessage(organization);
  $: bannerCTAHandler = new BannerCTAHandler(organization);
  $: ({ showStartTeamPlanDialog, startTeamPlanType, teamPlanEndDate } =
    bannerCTAHandler);

  function showBillingIssueBanner(message: BillingIssueMessage) {
    eventBus.emit("banner", {
      type: message.type,
      message: message.title + " " + message.description,
      iconType: message.iconType,
      ...(message.cta
        ? {
            cta: {
              type: "button",
              text: message.cta.text + "->",
              onClick() {
                return bannerCTAHandler.handle(message);
              },
            },
          }
        : {}),
    });
  }

  $: lastType = "default";
  $: if (!$billingIssueMessage.isFetching) {
    if ($billingIssueMessage.data) {
      showBillingIssueBanner($billingIssueMessage.data);
      if (
        lastType !== $billingIssueMessage.data.type &&
        $billingIssueMessage.data.type === "error"
      ) {
        // if issue changed to an error re-fetch projects to get hibernated projects
        void queryClient.refetchQueries(
          getAdminServiceListProjectsForOrganizationQueryKey(organization),
        );
      }
      lastType = $billingIssueMessage.data.type;
    } else {
      // when switching orgs we need to make sure we clear previous org's banner.
      // TODO: could this interfere with other banners?
      eventBus.emit("banner", {
        type: "clear",
        message: "",
        iconType: "none",
      });
    }
  }
</script>

<StartTeamPlanDialog
  bind:open={$showStartTeamPlanDialog}
  type={$startTeamPlanType}
  endDate={$teamPlanEndDate}
  {organization}
/>
