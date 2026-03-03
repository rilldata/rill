import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Chat", () => {
  test.describe.configure({ mode: "serial" });

  let testMessage: string;

  test("should send message and receive response", async ({ adminPage }) => {
    // Navigate to the chat page
    await adminPage.goto("/e2e/openrtb/-/ai");
    await expect(
      adminPage.getByText("How can I help you today?"),
    ).toBeVisible();

    // Send a message (with timestamp for uniqueness)
    const timestamp = Date.now();
    testMessage = `What happened recently? (test-${timestamp})`;
    await adminPage.getByRole("textbox").pressSequentially(testMessage);
    await adminPage.getByRole("button", { name: "Send" }).click();

    // Assert the response appears in the main chat area
    await expect(adminPage.getByText(`Echo: ${testMessage}`)).toBeVisible();

    // Assert our specific conversation appears in the sidebar conversation list
    await expect(
      adminPage
        .getByTestId("conversation-list")
        .getByTestId("conversation-item")
        .filter({ hasText: `test-${timestamp}` }),
    ).toBeVisible();

    // Assert "No conversations yet" is no longer visible
    const noConversationsElement = adminPage.getByTestId("no-conversations");
    if (await noConversationsElement.isVisible()) {
      await expect(noConversationsElement).not.toBeVisible();
    }
  });

  test("should submit positive feedback and display response", async ({
    adminPage,
  }) => {
    // Navigate to the chat page (conversation persists on server)
    await adminPage.goto("/e2e/openrtb/-/ai");

    // Click the conversation from the first test
    await adminPage
      .getByTestId("conversation-list")
      .getByTestId("conversation-item")
      .first()
      .click();

    // Wait for the assistant response to be visible
    await expect(adminPage.getByText(`Echo: ${testMessage}`)).toBeVisible();

    // Click the thumbs-up button on the assistant message
    const upvoteButton = adminPage.getByRole("button", {
      name: "Upvote response",
    });
    await upvoteButton.click();

    // The feedback response should be visible with the expected message
    await expect(
      adminPage.getByText("Thanks for the positive feedback!"),
    ).toBeVisible();

    // Refresh the page to test hydration
    await adminPage.reload();

    // Assert the feedback response is still visible after refresh
    await expect(
      adminPage.getByText("Thanks for the positive feedback!"),
    ).toBeVisible();

    // Assert the thumbs-up button is active
    await expect(
      adminPage.getByRole("button", { name: "Upvote response" }),
    ).toHaveAttribute("aria-pressed", "true");
  });
});
