/**
 * Types and constants for the AI feedback system
 */

export type FeedbackSentiment = "positive" | "negative";

export const FEEDBACK_CATEGORIES = [
  { id: "instruction_ignored", label: "Instruction ignored" },
  { id: "no_citation_links", label: "No citation links" },
  { id: "being_lazy", label: "Being lazy" },
  { id: "incorrect_information", label: "Incorrect information" },
  { id: "other", label: "Other" },
] as const;

export type FeedbackCategory = (typeof FEEDBACK_CATEGORIES)[number]["id"];
