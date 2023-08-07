import {
  extractFileExtension,
  extractTableName,
  sanitizeEntityName,
} from "@rilldata/web-common/features/sources/extract-table-name";
import { describe, it, expect } from "vitest";

function getVariations(
  fileName: string,
  expectedFileName: string,
  expectedExtension: string
) {
  return [
    `path/to/file/${fileName}`,
    `/path/to/file/${fileName}`,
    `./path/to/file/${fileName}`,
    fileName,
    `/${fileName}`,
    `./${fileName}`,
  ].map((path) => ({
    title: `fileName=${fileName} path=${path}`,
    expectedFileName,
    expectedExtension,
    path,
  }));
}
const Variations = [
  ...getVariations("22-02-10.parquet", "_22_02_10", ".parquet"),
  ...getVariations("-22-02-11.parquet", "_22_02_11", ".parquet"),
  ...getVariations("_22-02-12.parquet", "_22_02_12", ".parquet"),
  ...getVariations("table.parquet", "table", ".parquet"),
  ...getVariations("table.v1.parquet", "table_v1", ".v1.parquet"),
  ...getVariations("table", "table", ""),
  ...getVariations("table.v1.parquet.gz", "table_v1_parquet", ".v1.parquet.gz"),
];

describe("extract-table-name", () => {
  describe("should extract and sanitise table name", () => {
    for (const variation of Variations) {
      it(variation.title, () => {
        expect(sanitizeEntityName(extractTableName(variation.path))).toBe(
          variation.expectedFileName
        );
      });
    }
  });

  describe("should extract extension", () => {
    for (const variation of Variations) {
      it(variation.title, () => {
        expect(extractFileExtension(variation.path)).toBe(
          variation.expectedExtension
        );
      });
    }
  });
});
