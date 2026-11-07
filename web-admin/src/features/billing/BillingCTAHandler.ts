import { needsPaymentSetup } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
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
    // TODO: propagate errors
    switch (issueMessage.cta.type) {
      case "show-upgrade":
        this.showStartTeamPlanDialog.set(true);
        this.startTeamPlanType.set(
          issueMessage.cta.teamPlanDialogType ?? "base",
        );
        this.teamPlanEndDate.set(issueMessage.cta.teamPlanEndDate ?? "");
        break;

      case "payment": {
        const issues = await fetchOrganizationBillingIssues(this.organization);
        const setup = needsPaymentSetup(issues);
        window.open(
          await fetchPaymentsPortalURL(
            this.organization,
            window.location.href,
            setup,
          ),
          "_self",
        );
        break;
      }

      case "contact":
        window.Pylon("show");
        break;
    }
  }
}
