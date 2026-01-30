import { describe, it, expect } from "vitest";
import {
  getFileTypeFromPath,
  inferSourceName,
  compileSourceYAML,
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

describe("compileSourceYAML", () => {
  const mockConnector = {
    name: "duckdb",
    sourceProperties: [
      { key: "sql", type: 1 }, // TYPE_STRING = 1
    ],
  };

  it("should include create_secrets_from_connectors in output when provided", () => {
    const formValues = {
      sql: "SELECT * FROM read_parquet('gs://bucket/file.parquet')",
      create_secrets_from_connectors: "gcs_1",
    };

    const yaml = compileSourceYAML(mockConnector as any, formValues);

    expect(yaml).toContain("connector: duckdb");
    expect(yaml).toContain("create_secrets_from_connectors: gcs_1");
    expect(yaml).toContain("sql:");
  });

  it("should not include create_secrets_from_connectors when not provided", () => {
    const formValues = {
      sql: "SELECT * FROM read_parquet('gs://bucket/file.parquet')",
    };

    const yaml = compileSourceYAML(mockConnector as any, formValues);

    expect(yaml).toContain("connector: duckdb");
    expect(yaml).not.toContain("create_secrets_from_connectors");
  });

  it("should exclude empty string values", () => {
    const formValues = {
      sql: "SELECT * FROM table",
      create_secrets_from_connectors: "",
    };

    const yaml = compileSourceYAML(mockConnector as any, formValues);

    expect(yaml).not.toContain("create_secrets_from_connectors");
  });

  it("should exclude name field from output", () => {
    const formValues = {
      sql: "SELECT * FROM table",
      name: "my_model",
      create_secrets_from_connectors: "gcs_2",
    };

    const yaml = compileSourceYAML(mockConnector as any, formValues);

    expect(yaml).not.toContain("name: my_model");
    expect(yaml).toContain("create_secrets_from_connectors: gcs_2");
  });

  it("should use connector driver name for connector field", () => {
    const gcsConnector = {
      name: "gcs",
      sourceProperties: [],
    };

    const formValues = {
      path: "gs://bucket/file.csv",
    };

    const yaml = compileSourceYAML(gcsConnector as any, formValues);

    expect(yaml).toContain("connector: gcs");
  });

  it("should handle multiple form values correctly", () => {
    const formValues = {
      sql: "SELECT * FROM read_parquet('s3://bucket/file.parquet')",
      create_secrets_from_connectors: "s3_custom",
    };

    const yaml = compileSourceYAML(mockConnector as any, formValues);

    expect(yaml).toContain("type: model");
    expect(yaml).toContain("materialize: true");
    expect(yaml).toContain("connector: duckdb");
    expect(yaml).toContain("create_secrets_from_connectors: s3_custom");
  });
});
