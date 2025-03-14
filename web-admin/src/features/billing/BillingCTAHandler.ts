import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
import { wakeAllProjects } from "@rilldata/web-admin/features/organizations/hibernating/wakeAllProjects";
import {
  BillingBannerID,
  BillingBannerPriority,
} from "@rilldata/web-common/components/banner/constants";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { writable } from "svelte/store";

export class BillingCTAHandler {
  public showStartTeamPlanDialog = writable(false);
  public startTeamPlanType = writable<TeamPlanDialogTypes>("base");
  public teamPlanEndDate = writable("");
  public wakingProjects = writable(false);

  private static instances = new Map<string, BillingCTAHandler>();

  public constructor(private readonly organization: string) {}

  // maintain a cache of instances so that multiple components are in sync with internal state
  public static get(organization: string) {
    let instance: BillingCTAHandler;
    if (this.instances.has(organization)) {
      instance = this.instances.get(organization)!;
      instance.wakingProjects.set(false);
    } else {
      instance = new BillingCTAHandler(organization);
      this.instances.set(organization, instance);
    }
    return instance;
  }

  public async handle(issueMessage: BillingIssueMessage) {
    if (!issueMessage.cta) return;
    switch (issueMessage.cta.type) {
      case "upgrade":
        this.showStartTeamPlanDialog.set(true);
        this.startTeamPlanType.set(
          issueMessage.cta.teamPlanDialogType ?? "base",
        );
        this.teamPlanEndDate.set(issueMessage.cta.teamPlanEndDate ?? "");
        break;

      case "payment":
        window.open(
          await fetchPaymentsPortalURL(this.organization, window.location.href),
          "_self",
        );
        break;

      case "contact":
        window.Pylon("show");
        break;

      case "wake-projects":
        this.wakingProjects.set(true);
        eventBus.emit("add-banner", {
          id: BillingBannerID,
          priority: BillingBannerPriority,
          message: {
            type: "info",
            message: "Waking projects. We’ll notify you when they’re ready.",
            iconType: "loading",
          },
        });
        await wakeAllProjects(this.organization);
        this.wakingProjects.set(false);
        eventBus.emit("add-banner", {
          id: BillingBannerID,
          priority: BillingBannerPriority,
          message: {
            type: "success",
            message: "Your projects are awake and ready.",
            iconType: "check",
            cta: {
              type: "link",
              text: "View projects ->",
              url: `/${this.organization}`,
            },
          },
        });
        eventBus.emit("notification", {
          type: "success",
          message: "Projects are now ready and accessible",
        });
        break;
    }
  }
}
