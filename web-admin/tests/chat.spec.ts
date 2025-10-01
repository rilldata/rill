import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Chat", () => {
  test("should send message and receive response", async ({ adminPage }) => {
    // Navigate to the chat page
    await adminPage.goto("/e2e/openrtb/-/ai");
    await expect(
      adminPage.getByText("How can I help you today?"),
    ).toBeVisible();

    // Send a message (with timestamp for uniqueness)
    const timestamp = Date.now();
    const testMessage = `What happened recently? (test-${timestamp})`;
    await adminPage
      .getByPlaceholder("Ask about your data...")
      .fill(testMessage);
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
});
