import { expect, type Page } from "@playwright/test";
import type { EXAMPLES } from "@rilldata/web-common/features/welcome/constants";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";

/**
 * Resource status types from the runtime
 */
const ReconcileStatus = {
  IDLE: "RECONCILE_STATUS_IDLE",
  PENDING: "RECONCILE_STATUS_PENDING",
  RUNNING: "RECONCILE_STATUS_RUNNING",
} as const;

/**
 * Waits for all resources to reach idle state (reconciliation complete).
 * This polls the runtime API to check resource statuses.
 *
 * @param page - Playwright page
 * @param timeout - Maximum time to wait in ms (default 60s)
 */
export async function waitForAllResourcesIdle(
  page: Page,
  timeout = 60000,
): Promise<void> {
  const startTime = Date.now();

  const result = await asyncWaitUntil(
    async () => {
      // Get all resources via API
      const response = await page.request.get(
        "/v1/instances/default/resources",
      );

      if (!response.ok()) {
        return false;
      }

      const data = await response.json();
      const resources = data.resources ?? [];

      // Check if all resources are idle
      const allIdle = resources.every(
        (r: { meta?: { reconcileStatus?: string } }) =>
          r.meta?.reconcileStatus === ReconcileStatus.IDLE,
      );

      return allIdle;
    },
    timeout,
    1000, // Poll every 1 second
  );

  const elapsed = Date.now() - startTime;

  if (!result) {
    throw new Error(
      `Timeout: Not all resources reached idle state within ${timeout}ms`,
    );
  }

  console.log(`All resources reached idle state in ${elapsed}ms`);
}

/**
 * Gets all resource errors from the runtime.
 *
 * @param page - Playwright page
 * @returns Array of resources with errors
 */
export async function getResourceErrors(
  page: Page,
): Promise<Array<{ name: string; kind: string; error: string }>> {
  const response = await page.request.get("/v1/instances/default/resources");

  if (!response.ok()) {
    throw new Error(`Failed to get resources: ${response.status()}`);
  }

  const data = await response.json();
  const resources = data.resources ?? [];

  return resources
    .filter(
      (r: { meta?: { reconcileError?: string } }) => r.meta?.reconcileError,
    )
    .map(
      (r: {
        meta?: {
          name?: { name?: string; kind?: string };
          reconcileError?: string;
        };
      }) => ({
        name: r.meta?.name?.name ?? "unknown",
        kind: r.meta?.name?.kind ?? "unknown",
        error: r.meta?.reconcileError ?? "",
      }),
    );
}

/**
 * Waits for a specific resource to be idle with no errors.
 *
 * @param page - Playwright page
 * @param resourceName - Name of the resource
 * @param resourceKind - Kind of the resource (e.g., "rill.runtime.v1.MetricsView")
 * @param timeout - Maximum time to wait in ms (default 30s)
 */
export async function waitForResourceIdle(
  page: Page,
  resourceName: string,
  resourceKind: string,
  timeout = 30000,
): Promise<void> {
  const result = await asyncWaitUntil(
    async () => {
      const response = await page.request.get(
        `/v1/instances/default/resource?name.name=${encodeURIComponent(resourceName)}&name.kind=${encodeURIComponent(resourceKind)}`,
      );

      if (!response.ok()) {
        return false;
      }

      const data = await response.json();
      const resource = data.resource;

      return resource?.meta?.reconcileStatus === ReconcileStatus.IDLE;
    },
    timeout,
    500,
  );

  if (!result) {
    throw new Error(
      `Timeout: Resource ${resourceKind}/${resourceName} did not reach idle state within ${timeout}ms`,
    );
  }
}

/**
 * Checks that a metrics view has a valid spec (is not broken).
 *
 * @param page - Playwright page
 * @param metricsViewName - Name of the metrics view
 */
export async function assertMetricsViewValid(
  page: Page,
  metricsViewName: string,
): Promise<void> {
  const response = await page.request.get(
    `/v1/instances/default/resource?name.name=${encodeURIComponent(metricsViewName)}&name.kind=rill.runtime.v1.MetricsView`,
  );

  expect(response.ok()).toBe(true);

  const data = await response.json();
  const resource = data.resource;

  // Check that the metrics view has a valid spec
  expect(
    resource?.metricsView?.state?.validSpec,
    `MetricsView "${metricsViewName}" should have a valid spec`,
  ).toBeTruthy();

  // Check there's no reconcile error
  expect(
    resource?.meta?.reconcileError,
    `MetricsView "${metricsViewName}" should not have a reconcile error`,
  ).toBeFalsy();
}

/**
 * Checks that an explore has a valid spec.
 *
 * @param page - Playwright page
 * @param exploreName - Name of the explore
 */
export async function assertExploreValid(
  page: Page,
  exploreName: string,
): Promise<void> {
  const response = await page.request.get(
    `/v1/instances/default/resource?name.name=${encodeURIComponent(exploreName)}&name.kind=rill.runtime.v1.Explore`,
  );

  expect(response.ok()).toBe(true);

  const data = await response.json();
  const resource = data.resource;

  // Check that the explore has a valid spec
  expect(
    resource?.explore?.state?.validSpec,
    `Explore "${exploreName}" should have a valid spec`,
  ).toBeTruthy();

  // Check there's no reconcile error
  expect(
    resource?.meta?.reconcileError,
    `Explore "${exploreName}" should not have a reconcile error`,
  ).toBeFalsy();
}

/**
 * Checks that a model has been reconciled successfully.
 *
 * @param page - Playwright page
 * @param modelName - Name of the model
 */
export async function assertModelValid(
  page: Page,
  modelName: string,
): Promise<void> {
  const response = await page.request.get(
    `/v1/instances/default/resource?name.name=${encodeURIComponent(modelName)}&name.kind=rill.runtime.v1.Model`,
  );

  expect(response.ok()).toBe(true);

  const data = await response.json();
  const resource = data.resource;

  // Check that the model is idle
  expect(resource?.meta?.reconcileStatus).toBe(ReconcileStatus.IDLE);

  // Check there's no reconcile error
  expect(
    resource?.meta?.reconcileError,
    `Model "${modelName}" should not have a reconcile error`,
  ).toBeFalsy();

  // Check that the model has a result table
  expect(
    resource?.model?.state?.resultTable,
    `Model "${modelName}" should have a result table`,
  ).toBeTruthy();
}

/**
 * Asserts that the dashboard preview shows actual data and no errors.
 * This checks that the dashboard renders without showing the 404 error page.
 *
 * @param page - Playwright page
 */
export async function assertDashboardRendered(page: Page): Promise<void> {
  // Wait for the page to be fully loaded
  await page.waitForLoadState("networkidle");

  // Check that the 404 error page is NOT visible
  const errorPage = page.locator("text=Unable to load dashboard preview");
  await expect(errorPage).not.toBeVisible({ timeout: 5000 });

  // Check that dependency error message is NOT visible
  const dependencyError = page.locator("text=dependency error");
  await expect(dependencyError).not.toBeVisible({ timeout: 5000 });
}

/**
 * Asserts that explore dashboard data is visible (leaderboards have data).
 *
 * @param page - Playwright page
 */
export async function assertExploreHasData(page: Page): Promise<void> {
  // Wait for at least one leaderboard to be visible
  await expect(page.locator('[data-testid="leaderboard"]').first()).toBeVisible(
    { timeout: 10000 },
  );

  // Check that we have at least some data rows in the leaderboard
  const rows = page.locator('[data-testid="leaderboard"] tr');
  const rowCount = await rows.count();
  expect(rowCount).toBeGreaterThan(0);
}

/**
 * Extracts the first file dashboard name from an example project definition.
 *
 * @param example - Example project definition
 * @returns The dashboard/explore name derived from the firstFile path
 */
export function getDashboardNameFromExample(
  example: (typeof EXAMPLES)[number],
): string {
  // Extract name from path like "/dashboards/clickhouse_commits_explore.yaml"
  const match = example.firstFile.match(/\/dashboards\/(.+)\.yaml$/);
  if (match) {
    return match[1];
  }
  return example.firstFile
    .replace(/^\/dashboards\//, "")
    .replace(/\.yaml$/, "");
}

/**
 * Lists all resources of a specific kind.
 *
 * @param page - Playwright page
 * @param kind - Resource kind (e.g., "rill.runtime.v1.Model")
 * @returns Array of resource names
 */
export async function listResourcesOfKind(
  page: Page,
  kind: string,
): Promise<string[]> {
  const response = await page.request.get(
    `/v1/instances/default/resources?kind=${encodeURIComponent(kind)}`,
  );

  if (!response.ok()) {
    return [];
  }

  const data = await response.json();
  return (data.resources ?? []).map(
    (r: { meta?: { name?: { name?: string } } }) => r.meta?.name?.name ?? "",
  );
}

/**
 * Asserts that no critical resources have errors.
 * Critical resources include: Models, MetricsViews, Explores.
 *
 * @param page - Playwright page
 */
export async function assertNoCriticalErrors(page: Page): Promise<void> {
  const errors = await getResourceErrors(page);

  const criticalErrors = errors.filter(
    (e) =>
      e.kind.includes("Model") ||
      e.kind.includes("MetricsView") ||
      e.kind.includes("Explore") ||
      e.kind.includes("Canvas"),
  );

  if (criticalErrors.length > 0) {
    const errorMessages = criticalErrors
      .map((e) => `${e.kind}/${e.name}: ${e.error}`)
      .join("\n");
    throw new Error(`Critical resources have errors:\n${errorMessages}`);
  }
}
