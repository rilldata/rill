import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
import { writable } from "svelte/store";

export const showUpgradeDialog = writable(false);
export const upgradeDialogType = writable<TeamPlanDialogTypes>("base");
