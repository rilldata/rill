/**
 * Types and constants for the AI feedback system
 */

import { ToolName } from "../types";

export type FeedbackSentiment = "positive" | "negative";

/** Categories specific to the analyst agent */
const ANALYST_CATEGORIES = [
  { id: "incorrect_conclusions", label: "Incorrect conclusions" },
  { id: "incomplete_answer", label: "Incomplete answer" },
  { id: "wrong_data_used", label: "Wrong data used" },
  { id: "no_citations", label: "No citations" },
  { id: "other", label: "Other" },
] as const;

/** Categories specific to the developer agent */
const DEVELOPER_CATEGORIES = [
  { id: "code_doesnt_work", label: "Code doesn't work" },
  { id: "wrong_file_modified", label: "Wrong file modified" },
  { id: "didnt_complete_task", label: "Didn't complete task" },
  { id: "made_unwanted_changes", label: "Made unwanted changes" },
  { id: "other", label: "Other" },
] as const;

/** All possible feedback categories (union of all agent categories) */
export const FEEDBACK_CATEGORIES = [
  ...ANALYST_CATEGORIES,
  ...DEVELOPER_CATEGORIES,
] as const;

export type FeedbackCategory = (typeof FEEDBACK_CATEGORIES)[number]["id"];

const FEEDBACK_CATEGORIES_BY_AGENT: Record<
  string,
  typeof ANALYST_CATEGORIES | typeof DEVELOPER_CATEGORIES
> = {
  [ToolName.DEVELOPER_AGENT]: DEVELOPER_CATEGORIES,
  [ToolName.ANALYST_AGENT]: ANALYST_CATEGORIES,
};

/**
 * Get feedback categories applicable to the given agent.
 */
export function getCategoriesForAgent(agent: string) {
  return FEEDBACK_CATEGORIES_BY_AGENT[agent] ?? ANALYST_CATEGORIES;
}
