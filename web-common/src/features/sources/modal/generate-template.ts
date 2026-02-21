import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import {
  type V1GenerateTemplateResponse,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGenerateTemplate,
  runtimeServiceGetFile,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { replaceOrAddEnvVariable } from "../../connectors/code-utils";

/**
 * Call the GenerateTemplate RPC to produce YAML + env var names from
 * structured form data. The backend handles DuckDB rewrites, env var
 * naming/conflict resolution, and YAML formatting.
 */
export async function generateTemplate(
  instanceId: string,
  opts: {
    resourceType: string;
    driver: string;
    properties: Record<string, unknown>;
    connectorName?: string;
  },
): Promise<V1GenerateTemplateResponse> {
  return runtimeServiceGenerateTemplate(instanceId, {
    resourceType: opts.resourceType,
    driver: opts.driver,
    properties: opts.properties,
    connectorName: opts.connectorName,
  });
}

/**
 * Merge env vars returned by the GenerateTemplate RPC into the existing
 * `.env` file. The backend already resolved names and conflict suffixes,
 * so this is a straight key=value merge.
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
