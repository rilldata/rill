import { describe, it, expect } from "vitest";
import { getFileTypeFromPath, inferSourceName } from "./sourceUtils";

const gcsTests = [
  {
    path: "gs://bucket-name/folder-name/file-name.csv",
    expected: "file_name",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name.parquet",
    expected: "file_name",
    expectedExtension: "parquet",
  },
  {
    path: "gs://bucket-name/folder-name/file-name.csv.gz",
    expected: "file_name",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name",
    expected: "file_name",
    expectedExtension: "",
  },
  {
    path: "gs://bucket-name/folder-name/FILE-NAME.csv",
    expected: "FILE_NAME",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name123",
    expected: "file_name123",
    expectedExtension: "",
  },
];

const s3Tests = [
  {
    path: "s3://bucket-name/folder-name/file-name.csv",
    expected: "file_name",
  },
];

const httpTests = [
  {
    path: "http://example.com/folder-name/file-name.csv",
    expected: "file_name",
  },
  {
    path: "https://example.com/folder-name/file-name.csv",
    expected: "file_name",
  },
];

describe("inferSourceName", () => {
  // GCS
  it("should infer source name for GCS connector", () => {
    const connector = {
      name: "gcs",
    };

    gcsTests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });

  // S3
  it("should infer source name for S3 connector", () => {
    const connector = {
      name: "s3",
    };

    s3Tests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });

  // HTTPS
  it("should infer source name for HTTPS connector", () => {
    const connector = {
      name: "https",
    };

    httpTests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });
});

describe("getFileTypeFromPath", () => {
  it("should infer source name for given path", () => {
    gcsTests.forEach((test) => {
      const actual = getFileTypeFromPath(test.path);
      expect(actual).toEqual(test.expectedExtension);
    });
  });
});
