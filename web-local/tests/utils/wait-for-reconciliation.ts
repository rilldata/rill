import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";

interface Resource {
  meta?: {
    reconcileStatus?: string;
    reconcileError?: string;
    name?: { kind?: string; name?: string };
  };
}

/**
 * Waits for all resources in the instance to finish reconciling,
 * then asserts that none have errors.
 */
export async function waitForReconciliation(page: Page, timeoutMs = 60_000) {
  const baseUrl = new URL(page.url()).origin;
  let resources: Resource[] = [];

  const settled = await asyncWaitUntil(async () => {
    try {
      const response = await page.request.get(
        `${baseUrl}/v1/instances/default/resources`,
      );
      const body = await response.json();
      resources = body.resources ?? [];

      if (resources.length === 0) return false;

      // Exclude the ProjectParser — it's a meta-resource that stays
      // running while watching the repo and doesn't represent data errors.
      const dataResources = resources.filter(
        (r) => r.meta?.name?.kind !== "rill.runtime.v1.ProjectParser",
      );

      return dataResources.every(
        (r) => r.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE",
      );
    } catch {
      return false;
    }
  }, timeoutMs);

  const dataResources = resources.filter(
    (r) => r.meta?.name?.kind !== "rill.runtime.v1.ProjectParser",
  );

  if (!settled) {
    const pending = dataResources.filter(
      (r) => r.meta?.reconcileStatus !== "RECONCILE_STATUS_IDLE",
    );
    const details = pending
      .map(
        (r) =>
          `${r.meta?.name?.kind}/${r.meta?.name?.name}: ${r.meta?.reconcileStatus}`,
      )
      .join("\n");
    throw new Error(`Reconciliation timed out. Still pending:\n${details}`);
  }

  const errors = dataResources.filter((r) => r.meta?.reconcileError);
  if (errors.length > 0) {
    const errorDetails = errors
      .map(
        (r) =>
          `${r.meta?.name?.kind}/${r.meta?.name?.name}: ${r.meta?.reconcileError}`,
      )
      .join("\n");
    throw new Error(`Reconciliation errors:\n${errorDetails}`);
  }

  return dataResources;
}
