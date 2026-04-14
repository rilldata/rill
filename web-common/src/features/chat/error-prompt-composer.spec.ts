import { describe, it, expect } from "vitest";
import { composeErrorPrompt } from "./error-prompt-composer";

describe("composeErrorPrompt", () => {
  it("produces a minimal fix prompt with the file path", () => {
    const result = composeErrorPrompt("/models/my_model.sql");
    expect(result).toBe("Fix the errors in `/models/my_model.sql`");
  });
});
