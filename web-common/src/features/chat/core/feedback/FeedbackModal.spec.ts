import { fireEvent, render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import type { Conversation } from "../conversation";
import FeedbackModalHarness from "./__fixtures__/FeedbackModalHarness.svelte";

function renderModal() {
  const submitFeedback = vi.fn().mockResolvedValue(undefined);
  const conversation = { submitFeedback } as unknown as Conversation;
  render(FeedbackModalHarness, { props: { conversation } });
  return submitFeedback;
}

describe("FeedbackModal", () => {
  it("submits the rated message id, categories and comment", async () => {
    const submitFeedback = renderModal();

    await fireEvent.click(screen.getByText("Incorrect conclusions"));
    await fireEvent.input(screen.getByRole("textbox"), {
      target: { value: "wrong month" },
    });
    await fireEvent.click(screen.getByText("Submit"));

    expect(submitFeedback).toHaveBeenCalledWith(
      "assistant-message-1",
      "negative",
      ["incorrect_conclusions"],
      "wrong month",
    );
  });

  it("submits the rated message id on skip", async () => {
    const submitFeedback = renderModal();

    await fireEvent.click(screen.getByText("Skip"));

    expect(submitFeedback).toHaveBeenCalledWith(
      "assistant-message-1",
      "negative",
    );
  });
});
