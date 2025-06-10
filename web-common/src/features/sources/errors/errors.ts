import { SourceErrorCodes } from "../../../metrics/service/SourceEventTypes";
import { GRPC_ERROR_CODES } from "./constants";

export function hasDuckDBUnicodeError(message: string) {
  return message.match(
    /Invalid unicode \(byte sequence mismatch\) detected in CSV file./,
  );
}

const clickhouseErrorMap = {
  "connection refused":
    "Could not connect to ClickHouse server. Please check if the server is running and the host/port are correct.",
  "context deadline exceeded":
    "Connection to ClickHouse server timed out. Please check your network connection and server status.",
};

export function humanReadableErrorMessage(
  connectorName: string | undefined,
  code: number | undefined,
  message: string | undefined,
) {
  const unknownErrorStr = "An unknown error occurred.";

  const serverError = message;
  if (serverError === undefined) return unknownErrorStr;

  // gRPC error codes
  // https://pkg.go.dev/google.golang.org/grpc@v1.49.0/codes
  switch (code) {
    case GRPC_ERROR_CODES.Unknown: {
      // ClickHouse errors
      // source: errors reported by users
      if (connectorName === "clickhouse") {
        if (serverError.includes("connection refused")) {
          return clickhouseErrorMap["connection refused"];
        } else if (serverError.includes("context deadline exceeded")) {
          return clickhouseErrorMap["context deadline exceeded"];
        }
      }

      return serverError;
    }
    case GRPC_ERROR_CODES.InvalidArgument: {
      // Rill errors
      if (
        serverError.match(/an existing object with name '.*' already exists/)
      ) {
        return "A source with this name already exists. Please choose a different name.";
      }

      // AWS errors (ref: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
      if (connectorName === "s3") {
        if (serverError.includes("MissingRegion")) {
          return "Region not detected. Please enter a region.";
        } else if (serverError.includes("NoCredentialProviders")) {
          return "No credentials found. Please see the docs for how to configure AWS credentials.";
        } else if (serverError.includes("InvalidAccessKey")) {
          return "Invalid AWS access key. Please check your credentials.";
        } else if (serverError.includes("SignatureDoesNotMatch")) {
          return "Invalid AWS secret key. Please check your credentials.";
        } else if (serverError.includes("BucketRegionError")) {
          return "Bucket is not in the provided region. Please check your region.";
        } else if (serverError.includes("AccessDenied")) {
          return "Access denied. Please ensure you have the correct permissions.";
        } else if (serverError.includes("NoSuchKey")) {
          return "Invalid path. Please check your path.";
        } else if (serverError.includes("NoSuchBucket")) {
          return "Invalid bucket. Please check your bucket name.";
        } else if (serverError.includes("AuthorizationHeaderMalformed")) {
          return "Invalid authorization header. Please check your credentials.";
        }
      }

      // GCP errors (ref: https://cloud.google.com/storage/docs/json_api/v1/status-codes)
      if (connectorName === "gcs") {
        if (serverError.includes("could not find default credentials")) {
          return "No credentials found. Please see the docs for how to configure GCP credentials.";
        } else if (serverError.includes("Unauthorized")) {
          return "Unauthorized. Please check your credentials.";
        } else if (serverError.includes("AccessDenied")) {
          return "Access denied. Please ensure you have the correct permissions.";
        } else if (serverError.includes("object doesn't exist")) {
          return "Invalid path. Please check your path.";
        }
      }

      if (connectorName === "https") {
        if (serverError.includes("invalid file")) {
          return "The provided URL does not appear to have a valid dataset. Please check your path and try again.";
        } else if (serverError.includes("failed to fetch url")) {
          return "We could not connect to the provided URL. Please check your path and try again.";
        } else if (serverError.includes("file type not supported")) {
          return "Provided " + serverError;
        }
      }

      // DuckDB errors
      if (serverError.match(/expected \d* values per row, but got \d*/)) {
        return "Malformed CSV file: number of columns does not match header.";
      }

      // Fallback to raw server error
      return serverError;
    }
    case GRPC_ERROR_CODES.DeadlineExceeded: {
      return "The request timed out. Please ensure your service is running and try again.";
    }
    default:
      return unknownErrorStr;
  }
}

const errorTelemetryMap = {
  // AWS Errors
  missingRegion: SourceErrorCodes.MissingRegion,
  noCredentialProviders: SourceErrorCodes.Unauthorized,
  invalidAccessKey: SourceErrorCodes.InvalidAccessKey,
  signatureDoesNotMatch: SourceErrorCodes.SignatureDoesntMatch,
  bucketRegionError: SourceErrorCodes.BucketRegionError,
  accessDenied: SourceErrorCodes.AccessForbidden,
  noSuchKey: SourceErrorCodes.NoSuchKey,
  noSuchBucket: SourceErrorCodes.NoSuchBucket,
  authorizationHeaderMalformed: SourceErrorCodes.MalformedHeader,
  // GCP Errors
  "could not find default credentials": SourceErrorCodes.Unauthorized,
  NotFound: SourceErrorCodes.URLBroken,
  Unauthorized: SourceErrorCodes.Unauthorized,
  AccessDenied: SourceErrorCodes.AccessForbidden,
  PermissionDenied: SourceErrorCodes.AccessForbidden,
  "object doesn't exist": SourceErrorCodes.URLBroken,
  "no files found": SourceErrorCodes.URLBroken,
  // HTTPS Errors
  "Conversion error": SourceErrorCodes.MismatchedSchema,
  "Invalid Input Error": SourceErrorCodes.MismatchedSchema,
  "invalid file": SourceErrorCodes.MismatchedSchema,
  "failed to fetch url": SourceErrorCodes.URLBroken,
  "file type not supported": SourceErrorCodes.UnsupportedFileType,
  // Runtime errors
  "context deadline exceeded": SourceErrorCodes.RuntimeError,
  timeout: SourceErrorCodes.RuntimeError,
};

export function categorizeSourceError(errorMessage: string) {
  // check for connector errors
  for (const [key, value] of Object.entries(errorTelemetryMap)) {
    if (errorMessage.includes(key)) {
      return value;
    }
  }

  // check for duckdb errors
  if (errorMessage.match(/expected \d* values per row, but got \d*/)) {
    return SourceErrorCodes.MismatchedSchema;
  } else if (
    errorMessage.match(/Catalog Error: Table with name .* does not exist/)
  ) {
    return SourceErrorCodes.RuntimeError;
  } else if (hasDuckDBUnicodeError(errorMessage)) {
    return SourceErrorCodes.UnicodeError;
  }

  return SourceErrorCodes.Uncategorized;
}
