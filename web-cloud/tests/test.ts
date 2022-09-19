import { expect, test } from "@playwright/test";

test("empty test", async (_) => {
  // TODO: Run server-cloud behind the scenes to properly test this module
  expect("hello world").toBe("hello world");
});
