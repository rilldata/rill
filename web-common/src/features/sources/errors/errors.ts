import { SourceErrorCodes } from "../../../metrics/service/SourceEventTypes";
import { GRPC_ERROR_CODES } from "./constants";

export function hasDuckDBUnicodeError(message: string) {
  return message.match(
    /Invalid unicode \(byte sequence mismatch\) detected in CSV file./,
  );
}

const connectorErrorMap = {
  clickhouse: {
    "connection refused":
      "Could not connect to ClickHouse server. Please check if the server is running and the host/port are correct.",
    "context deadline exceeded":
      "Connection to ClickHouse server timed out. Please check your network connection and server status.",
  },
  // AWS errors (ref: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
  s3: {
    MissingRegion: "Region not detected. Please enter a region.",
    NoCredentialProviders:
      "No credentials found. Please see the docs for how to configure AWS credentials.",
    InvalidAccessKey: "Invalid AWS access key. Please check your credentials.",
    SignatureDoesNotMatch:
      "Invalid AWS secret key. Please check your credentials.",
    BucketRegionError:
      "Bucket is not in the provided region. Please check your region.",
    AccessDenied:
      "Access denied. Please ensure you have the correct permissions.",
    NoSuchKey: "Invalid path. Please check your path.",
    NoSuchBucket: "Invalid bucket. Please check your bucket name.",
    AuthorizationHeaderMalformed:
      "Invalid authorization header. Please check your credentials.",
  },
  // GCP errors (ref: https://cloud.google.com/storage/docs/json_api/v1/status-codes)
  gcs: {
    "could not find default credentials":
      "No credentials found. Please see the docs for how to configure GCP credentials.",
    Unauthorized: "Unauthorized. Please check your credentials.",
    AccessDenied:
      "Access denied. Please ensure you have the correct permissions.",
    "object doesn't exist": "Invalid path. Please check your path.",
  },
  https: {
    "invalid file":
      "The provided URL does not appear to have a valid dataset. Please check your path and try again.",
    "failed to fetch url":
      "We could not connect to the provided URL. Please check your path and try again.",
    "file type not supported": "The provided file type is not supported.",
  },
};

const CONNECTORS_WITH_CUSTOM_HANDLING = new Set(Object.keys(connectorErrorMap));

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

const DEFAULT_CONNECTOR_ERROR_TEMPLATE =
  "We received the following error when trying to connect to CONNECTOR_NAME. Please check your connection details and user permissions, then try again.";

export function humanReadableErrorMessage(
  connectorName: string | undefined,
  code: number | undefined,
  message: string | undefined,
) {
  const unknownErrorStr = "An unknown error occurred.";

  const serverError = message;
  if (serverError === undefined) return unknownErrorStr;

  // For connectors without custom handling, use the generic template
  if (connectorName && !CONNECTORS_WITH_CUSTOM_HANDLING.has(connectorName)) {
    return DEFAULT_CONNECTOR_ERROR_TEMPLATE.replace(
      "CONNECTOR_NAME",
      connectorName,
    );
  }

  // gRPC error codes
  // https://pkg.go.dev/google.golang.org/grpc@v1.49.0/codes
  switch (code) {
    case GRPC_ERROR_CODES.Unknown: {
      // ClickHouse errors
      if (connectorName === "clickhouse") {
        for (const [key, value] of Object.entries(
          connectorErrorMap.clickhouse,
        )) {
          if (serverError.includes(key)) {
            return value;
          }
        }
      }
      // For custom-handled connectors with no matching error, use the default template
      if (connectorName && CONNECTORS_WITH_CUSTOM_HANDLING.has(connectorName)) {
        return DEFAULT_CONNECTOR_ERROR_TEMPLATE.replace(
          "CONNECTOR_NAME",
          connectorName,
        );
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

      // Handle connector-specific errors
      if (connectorName && connectorName in connectorErrorMap) {
        const errorMap =
          connectorErrorMap[connectorName as keyof typeof connectorErrorMap];
        for (const [key, value] of Object.entries(errorMap)) {
          if (serverError.includes(key)) {
            return value;
          }
        }
        // For custom-handled connectors with no matching error, use the default template
        return DEFAULT_CONNECTOR_ERROR_TEMPLATE.replace(
          "CONNECTOR_NAME",
          connectorName,
        );
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
      // For custom-handled connectors with other error codes, use the default template
      if (connectorName && CONNECTORS_WITH_CUSTOM_HANDLING.has(connectorName)) {
        return DEFAULT_CONNECTOR_ERROR_TEMPLATE.replace(
          "CONNECTOR_NAME",
          connectorName,
        );
      }
      return unknownErrorStr;
  }
}

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
