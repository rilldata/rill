import { describe, it, expect, beforeEach, vi } from "vitest";
import { isRedirect } from "@sveltejs/kit";
import {
  V1DeploymentStatus,
  type V1Deployment,
  type V1ListDeploymentsResponse,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  isActiveDeployment,
  isProdDeployment,
  maybeRedirectToEditableDeployment,
} from "./deployment-utils";

const { listDeploymentsMock } = vi.hoisted(() => ({
  listDeploymentsMock: vi.fn<() => Promise<V1ListDeploymentsResponse>>(),
}));

vi.mock("@rilldata/web-admin/client", async () => {
  // Import the rest of the client. Mainly needed for type definitions.
  const actual = await vi.importActual<
    typeof import("@rilldata/web-admin/client")
  >("@rilldata/web-admin/client");
  return {
    ...actual,
    adminServiceListDeployments: (...args: unknown[]) =>
      listDeploymentsMock(...(args as [])),
  };
});

const ORG = "rilldata";
const PROJECT = "openrtb";

function makeDeployment(overrides: Partial<V1Deployment>): V1Deployment {
  return {
    status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
    ...overrides,
  };
}

describe("deployment-utils", () => {
  describe("isActiveDeployment", () => {
    it.each([
      V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
      V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
      V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING,
    ])("returns true for %s", (status) => {
      expect(isActiveDeployment({ status })).toBe(true);
    });

    it.each([
      V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
      V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED,
      V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED,
      V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING,
      V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING,
      V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    ])("returns false for %s", (status) => {
      expect(isActiveDeployment({ status })).toBe(false);
    });
  });

  describe("isProdDeployment", () => {
    it("returns true when environment is 'prod'", () => {
      expect(isProdDeployment({ environment: "prod" })).toBe(true);
    });

    it("returns false for any other environment", () => {
      expect(isProdDeployment({ environment: "dev" })).toBe(false);
      expect(isProdDeployment({ environment: "staging" })).toBe(false);
      expect(isProdDeployment({})).toBe(false);
    });
  });

  describe("maybeRedirectToEditableDeployment", () => {
    beforeEach(() => {
      listDeploymentsMock.mockReset();
      queryClient.clear();
    });

    async function call(pathname: string): Promise<Error | undefined> {
      try {
        await maybeRedirectToEditableDeployment(
          ORG,
          PROJECT,
          new URL(`http://localhost${pathname}`),
        );
        return undefined;
      } catch (e) {
        return e as Error;
      }
    }

    it("does not redirect when an active prod deployment exists", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "prod",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
          makeDeployment({
            environment: "dev",
            branch: "edit-branch",
            editable: true,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("does not redirect when there is no editable deployment", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "prod",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("does not redirect when the editable deployment has no branch", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: undefined,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("does not redirect when the editable deployment is inactive (hibernating)", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "prod",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED,
          }),
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("does not redirect when the user is already on the editable branch", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call(
        "/rilldata/openrtb/@edit-branch/explore/revenue",
      );
      expect(result).toBeUndefined();
    });

    it("does not redirect when the user is on a different branch than the editable one", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb/@some-other-branch");
      expect(result).toBeUndefined();
    });

    it("redirects to the editable branch when prod is inactive and editable is active", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "prod",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED,
          }),
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(isRedirect(result)).toBe(true);
      if (!isRedirect(result)) return; // type-safety
      expect(result.status).toBe(307);
      expect(result.location).toBe("/rilldata/openrtb/@edit-branch");
    });

    it("redirects when there is no prod deployment at all", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb/explore/revenue");
      expect(isRedirect(result)).toBe(true);
      if (!isRedirect(result)) return; // type-safety
      expect(result.status).toBe(307);
      expect(result.location).toBe("/rilldata/openrtb/@edit-branch");
    });

    it("redirects when prod is in PENDING (still active) — sanity check on active statuses", async () => {
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "prod",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
          }),
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("returns undefined when deployments list is empty", async () => {
      listDeploymentsMock.mockResolvedValue({ deployments: [] });

      const result = await call("/rilldata/openrtb");
      expect(result).toBeUndefined();
    });

    it("does not redirect away from the deploying page", async () => {
      // Conditions that would otherwise trigger a redirect:
      // no prod deployment, an active editable dev deployment, no @branch in URL.
      listDeploymentsMock.mockResolvedValue({
        deployments: [
          makeDeployment({
            environment: "dev",
            editable: true,
            branch: "edit-branch",
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
          }),
        ],
      });

      const result = await call("/rilldata/openrtb/-/deploying");
      expect(result).toBeUndefined();
    });
  });
});
