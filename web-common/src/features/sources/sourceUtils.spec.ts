import { describe, it, expect } from "vitest";
import {
  getFileTypeFromPath,
  inferSourceName,
  buildDuckDbQuery,
} from "./sourceUtils";

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

describe("buildDuckDbQuery", () => {
  const cases: Array<{
    name: string;
    path: string | undefined;
    options?: { defaultToJson?: boolean };
    expected: string;
  }> = [
    {
      name: "csv",
      path: "s3://bucket/data.csv",
      expected:
        "select * from read_csv('s3://bucket/data.csv', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "tsv",
      path: "s3://bucket/data.tsv",
      expected:
        "select * from read_csv('s3://bucket/data.tsv', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "txt",
      path: "/data/file.txt",
      expected:
        "select * from read_csv('/data/file.txt', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "parquet",
      path: "s3://bucket/data.parquet",
      expected: "select * from read_parquet('s3://bucket/data.parquet')",
    },
    {
      name: "json",
      path: "gs://bucket/data.json",
      expected:
        "select * from read_json('gs://bucket/data.json', auto_detect=true, format='auto')",
    },
    {
      name: "ndjson",
      path: "gs://bucket/data.ndjson",
      expected:
        "select * from read_json('gs://bucket/data.ndjson', auto_detect=true, format='auto')",
    },
    {
      name: "compound extension .v1.parquet.gz",
      path: "s3://bucket/data.v1.parquet.gz",
      expected: "select * from read_parquet('s3://bucket/data.v1.parquet.gz')",
    },
    {
      name: "compound extension .csv.gz",
      path: "s3://bucket/data.csv.gz",
      expected:
        "select * from read_csv('s3://bucket/data.csv.gz', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "unknown extension without defaultToJson",
      path: "s3://bucket/data.avro",
      expected: "select * from 's3://bucket/data.avro'",
    },
    {
      name: "unknown extension with defaultToJson",
      path: "https://api.example.com/data",
      options: { defaultToJson: true },
      expected:
        "select * from read_json('https://api.example.com/data', auto_detect=true, format='auto')",
    },
    {
      name: "undefined path",
      path: undefined,
      expected: "select * from ''",
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(buildDuckDbQuery(tc.path, tc.options)).toBe(tc.expected);
    });
  }
});
