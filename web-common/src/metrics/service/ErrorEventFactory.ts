import { MetricsEventFactory } from "./MetricsEventFactory";
import type {
  CommonFields,
  CommonUserFields,
  MetricsEvent,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "./MetricsTypes";

export enum ErrorEventAction {
  SourceError = "source-error",
}

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

export interface ErrorEvent extends MetricsEvent {
  action: ErrorEventAction;
  error_code: SourceErrorCodes;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  file_type: SourceFileType;
  connection_type: SourceConnectionType;
}

export class ErrorEventFactory extends MetricsEventFactory {
  public sourceErrorEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    error_code: SourceErrorCodes,
    connection_type: SourceConnectionType,
    file_type: SourceFileType
  ): ErrorEvent {
    const event = this.getBaseMetricsEvent(
      "error",
      commonFields,
      commonUserFields
    ) as ErrorEvent;
    event.action = ErrorEventAction.SourceError;
    event.space = space;
    event.screen_name = screen_name;
    event.error_code = error_code;
    event.connection_type = connection_type;
    event.file_type = file_type;
    return event;
  }
}
