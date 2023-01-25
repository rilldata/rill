import type { V1Connector } from "@rilldata/web-common/runtime-client";
import { sanitizeEntityName } from "./extract-table-name";

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

  let fileName = slug.split(".").shift();

  if (!fileName) return;

  fileName = fileName.replace(/\*/g, "");
  return sanitizeEntityName(fileName);
}
