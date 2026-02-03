/**
 * Types and constants for the AI feedback system
 */

import { ToolName } from "../types";

export type FeedbackSentiment = "positive" | "negative";

export const FEEDBACK_CATEGORIES = [
  { id: "instruction_ignored", label: "Instruction ignored" },
  { id: "no_citation_links", label: "No citation links" },
  { id: "being_lazy", label: "Being lazy" },
  { id: "incorrect_information", label: "Incorrect information" },
  { id: "other", label: "Other" },
] as const;

export type FeedbackCategory = (typeof FEEDBACK_CATEGORIES)[number]["id"];

/** Categories that only apply to specific agents */
const ANALYST_ONLY_CATEGORIES: FeedbackCategory[] = ["no_citation_links"];

/**
 * Get feedback categories applicable to the given agent.
 * Some categories (like "no citation links") only make sense for certain agents.
 */
export function getCategoriesForAgent(agent: string) {
  if (agent === ToolName.DEVELOPER_AGENT) {
    return FEEDBACK_CATEGORIES.filter(
      (c) => !ANALYST_ONLY_CATEGORIES.includes(c.id),
    );
  }
  return FEEDBACK_CATEGORIES;
}
