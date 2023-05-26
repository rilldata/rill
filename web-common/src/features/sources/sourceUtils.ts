import type { V1Connector } from "@rilldata/web-common/runtime-client";
import { sanitizeEntityName } from "./extract-table-name";
import { SourceErrorCodes } from "../../metrics/service/SourceEventTypes";
import { errorEvent } from "../../metrics/initMetrics";

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string
) {
  const topLineComment = `# Visit https://docs.rilldata.com/ to learn more about Rill code artifacts.`;

  if (connectorName !== "local_file") {
    values.uri = values.path;
    delete values.path;
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `${topLineComment}\n\ntype: "${connectorName}"\n` + compiledKeyValues;
}

export function inferSourceName(connector: V1Connector, path: string) {
  if (
    !path ||
    path.endsWith("/") ||
    (connector.name === "gcs" && !path.startsWith("gs://")) ||
    (connector.name === "s3" && !path.startsWith("s3://")) ||
    (connector.name === "https" &&
      !path.startsWith("https://") &&
      !path.startsWith("http://"))
  )
    return;

  const slug = path
    .split("/")
    .filter((s: string) => s.length > 0)
    .pop();

  if (!slug) return;

  const fileName = slug.split(".").shift();

  if (!fileName) return;

  return sanitizeEntityName(fileName);
}

export function sourceErrorTelemetryHandler(
  space,
  screenName,
  errors,
  connectionType,
  fileName
) {
  const categorizedError = parseSourceError(errors);
  const fileType = getFileTypeFromName(fileName);

  // errorEvent.fireSourceErrorEvent(
  //   space,
  //   screenName,
  //   categorizedError,
  //   connectionType,
  //   fileType
  // );
}

function getFileTypeFromName(fileName) {
  const fileType = fileName.split(".").pop();

  if (!fileType) return "";

  if (fileType === "gz") {
    return fileName.split(".").slice(-2).shift();
  }

  return fileType;
}

export function parseSourceError(errors) {
  console.log(errors);

  for (const error of errors) {
    if (error?.message?.includes("Invalid Error: Invalid Input Error")) {
      return SourceErrorCodes.InvalidInput;
    }
    if (error?.message?.includes("PermissionDenied")) {
      return SourceErrorCodes.AccessForbidden;
    }
    if (error?.message?.includes("failed to fetch url")) {
      return SourceErrorCodes.URLBroken;
    }
    if (error?.message?.includes("file type not supported")) {
      return SourceErrorCodes.UnsupportedFileType;
    }
  }

  return SourceErrorCodes.Uncategorized;
}
