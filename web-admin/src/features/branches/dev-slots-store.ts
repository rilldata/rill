import { writable } from "svelte/store";

// TODO: Remove this store once the deployment API supports per-deployment dev_slots.
// Currently dev_slots is not in UpdateProjectRequest; this store provides a local
// UI-only override shared between DeploymentsPage and BranchesSection.
// When the API lands, replace this with a real mutation in ManageSlotsModal (slotType="dev").
export const devSlotsOverride = writable<number | null>(null);
