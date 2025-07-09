import { writable } from "svelte/store";

/**
 * A store that tracks whether a child component has requested tooltip suppression.
 * This enables us to disentangle the tooltip state in certain cases where it doesn't
 * make sense to have the user deal with the logic of suppression.
 */
export const childRequestedTooltipSuppression = writable(false);

export const CHILD_REQUESTED_TOOLTIP_SUPPRESSION_CONTEXT_KEY =
  "rill:app:childRequestedTooltipSuppression";
