import { describe, it, expect } from "vitest";
import {
  detectConnectorFromPath,
  detectConnectorFromContent,
  detectConnector,
  deriveConnectorType,
} from "./connector-type-detector";

describe("connector-type-detector", () => {
  describe("detectConnectorFromPath", () => {
    it("should return undefined for null/undefined/empty inputs", () => {
      expect(detectConnectorFromPath(null)).toBeUndefined();
      expect(detectConnectorFromPath(undefined)).toBeUndefined();
      expect(detectConnectorFromPath("")).toBeUndefined();
    });

    it("should detect S3 paths", () => {
      expect(detectConnectorFromPath("s3://bucket/file.parquet")).toBe("s3");
      expect(detectConnectorFromPath("s3a://bucket/file.csv")).toBe("s3");
    });

    it("should detect GCS paths", () => {
      expect(detectConnectorFromPath("gs://bucket/file.parquet")).toBe("gcs");
      expect(detectConnectorFromPath("gcs://bucket/file.csv")).toBe("gcs");
    });

    it("should detect Azure paths", () => {
      expect(detectConnectorFromPath("azure://container/file.json")).toBe(
        "azure",
      );
      expect(detectConnectorFromPath("az://container/file.parquet")).toBe(
        "azure",
      );
      expect(detectConnectorFromPath("abfs://container/file.csv")).toBe(
        "azure",
      );
      expect(
        detectConnectorFromPath("abfss://container@account/file.csv"),
      ).toBe("azure");
    });

    it("should detect HTTPS paths", () => {
      expect(detectConnectorFromPath("https://example.com/data.csv")).toBe(
        "https",
      );
      expect(detectConnectorFromPath("http://example.com/data.csv")).toBe(
        "https",
      );
    });

    it("should be case-insensitive", () => {
      expect(detectConnectorFromPath("S3://bucket/file.parquet")).toBe("s3");
      expect(detectConnectorFromPath("GS://bucket/file.csv")).toBe("gcs");
      expect(detectConnectorFromPath("HTTPS://example.com/data")).toBe("https");
      expect(detectConnectorFromPath("Azure://container/file")).toBe("azure");
    });

    it("should return undefined for local file paths", () => {
      expect(detectConnectorFromPath("/local/file.csv")).toBeUndefined();
      expect(
        detectConnectorFromPath("./relative/file.parquet"),
      ).toBeUndefined();
      expect(detectConnectorFromPath("file.csv")).toBeUndefined();
    });

    it("should return undefined for unknown protocols", () => {
      expect(detectConnectorFromPath("ftp://server/file.csv")).toBeUndefined();
      expect(
        detectConnectorFromPath("custom://bucket/file.parquet"),
      ).toBeUndefined();
    });
  });

  describe("detectConnectorFromContent", () => {
    it("should return undefined for null/undefined/empty inputs", () => {
      expect(detectConnectorFromContent(null)).toBeUndefined();
      expect(detectConnectorFromContent(undefined)).toBeUndefined();
      expect(detectConnectorFromContent("")).toBeUndefined();
    });

    it("should detect S3 URLs embedded in SQL", () => {
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_parquet('s3://bucket/file.parquet')",
        ),
      ).toBe("s3");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_csv('s3a://bucket/data.csv')",
        ),
      ).toBe("s3");
    });

    it("should detect GCS URLs embedded in SQL", () => {
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_parquet('gs://bucket/file.parquet')",
        ),
      ).toBe("gcs");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_json('gcs://bucket/data.json')",
        ),
      ).toBe("gcs");
    });

    it("should detect Azure URLs embedded in SQL", () => {
      expect(
        detectConnectorFromContent(
          "SELECT * FROM 'azure://container/file.parquet'",
        ),
      ).toBe("azure");
      expect(
        detectConnectorFromContent("SELECT * FROM 'az://container/data.csv'"),
      ).toBe("azure");
      expect(
        detectConnectorFromContent("SELECT * FROM 'abfs://container/data.csv'"),
      ).toBe("azure");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM 'abfss://container@account/data.csv'",
        ),
      ).toBe("azure");
    });

    it("should detect HTTPS URLs only when they contain data file extensions", () => {
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_csv('https://example.com/data.csv')",
        ),
      ).toBe("https");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_parquet('https://storage.example.com/file.parquet')",
        ),
      ).toBe("https");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM 'https://example.com/data.json'",
        ),
      ).toBe("https");
    });

    it("should NOT detect HTTPS URLs without data file extensions", () => {
      expect(
        detectConnectorFromContent(
          "-- see https://docs.example.com/guide for details",
        ),
      ).toBeUndefined();
      expect(
        detectConnectorFromContent(
          "SELECT * FROM my_table -- https://example.com/api",
        ),
      ).toBeUndefined();
    });

    it("should detect all supported data file extensions in HTTP URLs", () => {
      const extensions = [
        ".parquet",
        ".csv",
        ".json",
        ".ndjson",
        ".jsonl",
        ".xlsx",
        ".xls",
        ".tsv",
      ];
      for (const ext of extensions) {
        expect(
          detectConnectorFromContent(
            `SELECT * FROM 'https://example.com/data${ext}'`,
          ),
        ).toBe("https");
      }
    });

    it("should detect DuckDB read functions as local_file", () => {
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_parquet('/local/file.parquet')",
        ),
      ).toBe("local_file");
      expect(
        detectConnectorFromContent("SELECT * FROM read_csv('/local/file.csv')"),
      ).toBe("local_file");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_json('/local/file.json')",
        ),
      ).toBe("local_file");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_ndjson('/local/file.ndjson')",
        ),
      ).toBe("local_file");
    });

    it("should prioritize cloud storage over DuckDB read functions", () => {
      // S3 URL inside a read_parquet should return s3, not local_file
      expect(
        detectConnectorFromContent(
          "SELECT * FROM read_parquet('s3://bucket/file.parquet')",
        ),
      ).toBe("s3");
    });

    it("should return undefined for plain SQL without URLs or read functions", () => {
      expect(
        detectConnectorFromContent("SELECT * FROM other_model"),
      ).toBeUndefined();
      expect(
        detectConnectorFromContent(
          "SELECT a, b FROM table1 JOIN table2 ON a = b",
        ),
      ).toBeUndefined();
    });

    it("should be case-insensitive", () => {
      expect(
        detectConnectorFromContent("SELECT * FROM 'S3://bucket/file.parquet'"),
      ).toBe("s3");
      expect(
        detectConnectorFromContent(
          "SELECT * FROM READ_PARQUET('/local/file.parquet')",
        ),
      ).toBe("local_file");
    });
  });

  describe("detectConnector", () => {
    it("should prioritize path detection over content detection", () => {
      expect(
        detectConnector(
          "s3://bucket/file.parquet",
          "SELECT * FROM read_csv('/local/file.csv')",
        ),
      ).toBe("s3");
    });

    it("should fall back to content detection when path has no match", () => {
      expect(
        detectConnector(
          null,
          "SELECT * FROM read_parquet('gs://bucket/file.parquet')",
        ),
      ).toBe("gcs");
    });

    it("should return undefined when neither matches", () => {
      expect(detectConnector(null, "SELECT * FROM my_table")).toBeUndefined();
      expect(detectConnector(null, null)).toBeUndefined();
    });

    it("should handle path-only detection", () => {
      expect(detectConnector("gs://bucket/file.csv", null)).toBe("gcs");
    });
  });

  describe("deriveConnectorType", () => {
    it("should detect from partition resolver properties first", () => {
      expect(
        deriveConnectorType({
          partitionsResolverProperties: { uri: "s3://bucket/part/*.parquet" },
          sourcePath: "gs://other-bucket/file.csv",
          sqlContent: "SELECT * FROM read_csv('/local/file.csv')",
          inputConnector: "postgres",
        }),
      ).toBe("s3");
    });

    it("should skip non-string partition resolver values", () => {
      expect(
        deriveConnectorType({
          partitionsResolverProperties: { count: 42, flag: true },
          sourcePath: "gs://bucket/file.csv",
        }),
      ).toBe("gcs");
    });

    it("should fall back to source path when partitions have no match", () => {
      expect(
        deriveConnectorType({
          partitionsResolverProperties: { key: "no-cloud-prefix" },
          sourcePath: "azure://container/file.parquet",
        }),
      ).toBe("azure");
    });

    it("should fall back to SQL content when source path has no match", () => {
      expect(
        deriveConnectorType({
          sourcePath: "/local/path/model.sql",
          sqlContent: "SELECT * FROM read_parquet('gs://bucket/file.parquet')",
        }),
      ).toBe("gcs");
    });

    it("should fall back to inputConnector when nothing else matches", () => {
      expect(
        deriveConnectorType({
          sourcePath: "/models/my_model.sql",
          sqlContent: "SELECT * FROM other_model",
          inputConnector: "postgres",
        }),
      ).toBe("postgres");
    });

    it("should return undefined when nothing matches and no inputConnector", () => {
      expect(
        deriveConnectorType({
          sqlContent: "SELECT 1",
        }),
      ).toBeUndefined();
    });

    it("should return undefined for empty options", () => {
      expect(deriveConnectorType({})).toBeUndefined();
    });

    it("should handle null/undefined values gracefully", () => {
      expect(
        deriveConnectorType({
          partitionsResolverProperties: null,
          sourcePath: null,
          sqlContent: null,
          inputConnector: null,
        }),
      ).toBeUndefined();
    });

    it("should use sourcePath as content fallback when sqlContent is absent", () => {
      // sourcePath doesn't match a prefix but contains an embedded URL
      expect(
        deriveConnectorType({
          sourcePath: "read_parquet('s3://bucket/file.parquet')",
        }),
      ).toBe("s3");
    });
  });
});
