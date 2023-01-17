import {
  getEmbeddedReferences,
  getTableReferences,
} from "@rilldata/web-common/features/models/utils/get-table-references";
import { tests } from "@rilldata/web-common/features/models/utils/get-table-references/test-data";

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
