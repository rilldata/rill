import { expect, type Page } from "@playwright/test";

export async function validateYamlContents(
  page: Page,
  expectedSnippets: string[],
  notExpectedSnippets: string[] = [],
) {
  const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
  for (const snippet of expectedSnippets) {
    await expect(envEditor).toContainText(snippet);
  }
  for (const snippet of notExpectedSnippets) {
    await expect(envEditor).not.toContainText(snippet);
  }
}
