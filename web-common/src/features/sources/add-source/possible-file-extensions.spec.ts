import { describe, expect, it } from "vitest";
import { fileHasValidExtension } from "@rilldata/web-common/features/sources/add-source/possible-file-extensions";

describe("fileHasValidExtension", () => {
  describe("positive cases", () => {
    [
      "file.csv.gz",
      "file.csv",
      "/path/to/file.parquet",
      "/path/to/file.parquet.gz",
      "/path/to/../file.txt",
      "/path/to/../file.txt.gz",
      "https://server.com/path/file.json",
      "https://server.com/path/file.json.gz",
    ].forEach((positiveCase) => {
      it(`${positiveCase} => true`, () => {
        expect(fileHasValidExtension(positiveCase)).toBeTruthy();
      });
    });
  });

  describe("negative cases", () => {
    [
      "file",
      "file.gz", // just gz is not allowed
      "https://server.com/path/file.js",
    ].forEach((positiveCase) => {
      it(`${positiveCase} => false`, () => {
        expect(fileHasValidExtension(positiveCase)).toBeFalsy();
      });
    });
  });
});
