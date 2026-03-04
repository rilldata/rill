import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGenerateFile,
  runtimeServiceGetFile,
  runtimeServiceGetInstance,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { replaceOrAddEnvVariable } from "../../connectors/code-utils";
import { normalizeOlapForTemplate } from "./connector-schemas";
import { OLAP_ENGINES } from "./constants";

const OLAP_SET = new Set(OLAP_ENGINES);

// Cache OLAP per instance to avoid a network call on every YAML preview keystroke.
const olapCache = new Map<string, string>();

/**
 * Resolve the template name from (driver, olap).
 * OLAP engine connectors have standalone templates (e.g., "clickhouse").
 * Source connectors use combined templates (e.g., "s3-duckdb", "athena-duckdb")
 * regardless of whether we're rendering the connector or model output.
 */
function resolveTemplateName(driver: string, olap: string): string {
  if (OLAP_SET.has(driver)) return driver;
  return `${driver}-${olap}`;
}

/**
 * Call the GenerateFile RPC to produce YAML + env var names from
 * structured form data. The backend handles rewrites, env var
 * naming/conflict resolution, and YAML formatting via declarative templates.
 *
 * Uses preview mode so the server renders without writing files;
 * the caller handles file writing and reconciliation.
 */
export async function generateTemplate(
  instanceId: string,
  opts: {
    resourceType: string;
    driver: string;
    properties: Record<string, unknown>;
    connectorName?: string;
  },
): Promise<{ blob: string; envVars: Record<string, string> }> {
  // Resolve OLAP; cached per instance since it doesn't change during a session.
  let olap = "duckdb";
  if (!OLAP_SET.has(opts.driver)) {
    const cached = olapCache.get(instanceId);
    if (cached) {
      olap = cached;
    } else {
      const resp = await runtimeServiceGetInstance(instanceId, {
        sensitive: true,
      });
      olap = normalizeOlapForTemplate(resp.instance?.olapConnector ?? "duckdb");
      olapCache.set(instanceId, olap);
    }
  }

  const templateName = resolveTemplateName(opts.driver, olap);

  const response = await runtimeServiceGenerateFile(instanceId, {
    templateName,
    output: opts.resourceType,
    properties: opts.properties,
    connectorName: opts.connectorName,
    preview: true,
  });

  // Flatten to match the interface callers expect
  return {
    blob: response.files?.[0]?.blob ?? "",
    envVars: response.envVars ?? {},
  };
}

/**
 * Merge env vars returned by GenerateFile into the existing `.env` file.
 * The backend already resolved names and conflict suffixes, so this is
 * a straight key=value merge.
 *
 * Returns the updated blob and the original blob (for rollback).
 */
export async function mergeEnvVars(
  queryClient: QueryClient,
  envVars: Record<string, string>,
): Promise<{ newBlob: string; originalBlob: string }> {
  const instanceId = get(runtime).instanceId;

  // Invalidate cache to get fresh .env content
  await queryClient.invalidateQueries({
    queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
  });

  let blob: string;
  let originalBlob: string;
  try {
    const file = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    blob = file.blob || "";
    originalBlob = blob;
  } catch (error) {
    if ((error as any)?.response?.data?.message?.includes("no such file")) {
      blob = "";
      originalBlob = "";
    } else {
      throw error;
    }
  }

  for (const [key, value] of Object.entries(envVars)) {
    if (!key || !value) continue;
    blob = replaceOrAddEnvVariable(blob, key, value);
  }

  return { newBlob: blob, originalBlob };
}
