import { describe, expect, it } from "vitest";
import { getEmbeddedReferences, getTableReferences } from ".";
import { tests } from "./test-data";

describe("getAllTableReferences", () => {
  it("correctly assesses the table references", () => {
    tests.forEach((test) => {
      const references = getTableReferences(test.query);
      expect(references).toEqual(test.references);
    });
  });

  it("correctly assesses embedded table references", () => {
    tests.forEach((test) => {
      const references = getEmbeddedReferences(test.query);
      expect(references).toEqual(test.embeddedReferences);
    });
  });
});
