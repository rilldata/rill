import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";

export function useDashboardPolicyCheck(
  client: RuntimeClient,
  filePath: string,
) {
  return createRuntimeServiceGetFile(
    client,
    {
      path: filePath,
    },
    {
      query: {
        enabled: !!filePath,
        select: (data) => {
          if (!data.blob) return false;
          const yamlObj = parse(data.blob);
          const securityPolicy = yamlObj?.security;
          return !!securityPolicy || hasVisibilityConditions(yamlObj?.rows);
        },
      },
    },
  );
}

/**
 * Whether the canvas file shows some of its content to some users only, through `if` conditions.
 * Reading the file requires access to the project's files; without it, the check resolves to an error and no data.
 */
export function useVisibilityConditionsCheck(
  client: RuntimeClient,
  filePath: string,
) {
  return createRuntimeServiceGetFile(
    client,
    {
      path: filePath,
    },
    {
      query: {
        enabled: !!filePath,
        retry: false,
        select: (data) =>
          !!data.blob &&
          hasVisibilityConditions(
            (parse(data.blob) as { rows?: unknown } | null)?.rows,
          ),
      },
    },
  );
}

/**
 * Whether canvas rows (including tab groups and their tabs) use `if` conditions to show content to some users only.
 * Previewing as another user is then useful even without a security policy.
 */
export function hasVisibilityConditions(rows: unknown): boolean {
  if (!Array.isArray(rows)) return false;
  return rows.some((row) => {
    if (!row || typeof row !== "object") return false;
    if ("if" in row) return true;
    const { items, tabs } = row as { items?: unknown; tabs?: unknown };
    if (
      Array.isArray(items) &&
      items.some((item) => !!item && typeof item === "object" && "if" in item)
    ) {
      return true;
    }
    return (
      Array.isArray(tabs) &&
      tabs.some(
        (tab) =>
          !!tab &&
          typeof tab === "object" &&
          ("if" in tab ||
            hasVisibilityConditions((tab as { rows?: unknown }).rows)),
      )
    );
  });
}

export function useRillYamlPolicyCheck(client: RuntimeClient) {
  return createRuntimeServiceGetFile(
    client,
    {
      path: "rill.yaml",
    },
    {
      query: {
        select: (data) => {
          if (!data.blob) return false;
          const yamlObj = parse(data.blob);
          const exploresSecurityPolicy = yamlObj?.explores?.security;
          const metricsViewsSecurityPolicy = yamlObj?.metricsViews?.security;
          const canvasesSecurityPolicy = yamlObj?.canvases?.security;
          return (
            !!exploresSecurityPolicy ||
            !!metricsViewsSecurityPolicy ||
            !!canvasesSecurityPolicy
          );
        },
      },
    },
  );
}
