import { writable } from "svelte/store";
import type { ControlDraft } from "../control-edit";

/**
 * A request to open the control dialog for an item, prefilled. Set by the "Let viewers choose"
 * links next to a widget's fields; the item's viewer controls section picks it up.
 */
export const controlRequest = writable<{
  /** The item's YAML path, joined with "/". */
  itemPath: string;
  draft: ControlDraft;
} | null>(null);
