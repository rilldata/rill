import {
  extractFileExtension,
  extractTableName,
  sanitizeEntityName,
} from "@rilldata/web-common/features/sources/extract-table-name";
import { describe, it, expect } from "vitest";

describe("extract-table-name", () => {
  describe("should extract and sanitise table name", () => {
    for (const variation of Variations) {
      it(`${variation.title} path=${variation.path}`, () => {
        expect(sanitizeEntityName(extractTableName(variation.path))).toBe(
          variation.expectedFileName
        );
      });
    }
  });

  describe("should extract extension", () => {
    for (const variation of Variations) {
      it(`${variation.title} path=${variation.path}`, () => {
        expect(extractFileExtension(variation.path)).toBe(
          variation.expectedExtension
        );
      });
    }
  });
});

function getVariations(
  fileName,
  expectedFileName,
  expectedExtension = ".parquet"
) {
  const title = `fileName=${fileName}`;
  const variation = {
    title,
    expectedFileName,
    expectedExtension,
  };
  return [
    {
      path: `path/to/file/${fileName}`,
      ...variation,
    },
    {
      path: `/path/to/file/${fileName}`,
      ...variation,
    },
    {
      path: `./path/to/file/${fileName}`,
      ...variation,
    },
    {
      path: fileName,
      ...variation,
    },
    {
      path: `/${fileName}`,
      ...variation,
    },
    {
      path: `./${fileName}`,
      ...variation,
    },
  ];
}
const Variations = [
  ...getVariations("22-02-10.parquet", "_22_02_10"),
  ...getVariations("-22-02-11.parquet", "_22_02_11"),
  ...getVariations("_22-02-12.parquet", "_22_02_12"),
  ...getVariations("table.parquet", "table"),
  ...getVariations("table.v1.parquet", "table_v1", ".v1.parquet"),
  ...getVariations("table", "table", ""),
  ...getVariations("table.v1.parquet.gz", "table_v1_parquet", ".v1.parquet.gz"),
];
