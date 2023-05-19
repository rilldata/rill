export enum SourceErrorCodes {
  UnsupportedFileType = "unsupported_file_type",
  MismatchedSchema = "mismatched_schema",
  NoServerResponse = "no_server_response",
  ExceedDataSizeMemory = "exceed_data_size_memory",
  ExceedDataSizeRuntime = "exceed_data_size_runtime",
  AccessForbidden = "access_forbidden",
  Unauthorized = "Unauthorized",
  URLBroken = "url_broken",
}

export enum SourceFileType {
  CSV = "csv",
  NDJSON = "ndjson",
  Parquet = "parquet",
}

export enum SourceConnectionType {
  S3 = "s3",
  GCS = "gcs",
  Https = "https",
  Local = "local",
}

export enum SourceEventMedium {
  Button = "button",
  Card = "card",
}

export enum SourceBehaviourEventAction {
  SourceModal = "source-modal",
  SourceCancel = "source-cancel",
  SourceAdd = "source-add",
}
