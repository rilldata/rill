/**
 * 1. base  - When user chooses to upgrade from a trial plan.
 * 2. size  - When user hits the size limit and wants to upgrade.
 * 3. org   - When user hits the organization limit and wants to upgrade.
 * 4. proj  - When user hits the project limit and wants to upgrade.
 * 5. renew - After user cancels a subscription and wants to renew.
 * 6. trial-expired - After a trial has expired with grace period also ended.
 */
export type TeamPlanDialogTypes =
  | "base"
  | "size"
  | "org"
  | "proj"
  | "renew"
  | "trial-expired";
