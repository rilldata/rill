export enum SourceErrorCodes {
  InvalidInput = "invalid_input",
  UnsupportedFileType = "unsupported_file_type",
  MismatchedSchema = "mismatched_schema",
  RuntimeError = "runtime_error",
  NoServerResponse = "no_server_response",
  ExceedDataSizeMemory = "exceed_data_size_memory",
  ExceedDataSizeRuntime = "exceed_data_size_runtime",
  AccessForbidden = "access_forbidden",
  Unauthorized = "unauthorized",
  URLBroken = "url_broken",
  Uncategorized = "uncategorized",
  MissingRegion = "missing_region",
  InvalidAccessKey = "invalid_access_key",
  SignatureDoesntMatch = "signature_doesnt_match",
  BucketRegionError = "bucket_region_error",
  NoSuchKey = "no_such_key",
  NoSuchBucket = "no_such_bucket",
  MalformedHeader = "malformed_header",
  UnicodeError = "unicode_error",
}

export enum SourceFileType {
  CSV = "csv",
  NDJSON = "ndjson",
  Parquet = "parquet",
  JSON = "json",
  TSV = "tsv",
  TXT = "txt",
}

export enum SourceConnectionType {
  S3 = "s3",
  GCS = "gcs",
  Https = "https",
  Local = "local",
}

export enum SourceBehaviourEventAction {
  SourceModal = "source-modal",
  SourceCancel = "source-cancel",
  SourceAdd = "source-add",
}

export const connectorToSourceConnectionType = {
  s3: SourceConnectionType.S3,
  gcs: SourceConnectionType.GCS,
  https: SourceConnectionType.Https,
  local: SourceConnectionType.Local,
};
