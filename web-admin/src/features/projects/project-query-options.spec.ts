import {
  V1DeploymentStatus,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
import { describe, expect, it } from "vitest";
import { baseGetProjectQueryOptions } from "./project-query-options";

const refetchInterval = baseGetProjectQueryOptions.refetchInterval;

function poll(data: V1GetProjectResponse | undefined) {
  if (typeof refetchInterval !== "function") {
    throw new Error("expected refetchInterval to be a function");
  }
  return refetchInterval({
    state: { data },
  } as unknown as Parameters<typeof refetchInterval>[0]);
}

describe("baseGetProjectQueryOptions.refetchInterval", () => {
  it("polls while a loaded project is hibernating (no deployment)", () => {
    expect(poll({ project: { name: "p" } })).toBe(2000);
  });

  it("polls quickly while the deployment is pending", () => {
    expect(
      poll({
        project: { name: "p" },
        deployment: { status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING },
      }),
    ).toBe(1000);
  });

  it("refetches the JWT proactively while running", () => {
    expect(
      poll({
        project: { name: "p" },
        deployment: { status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING },
      }),
    ).toBe(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2);
  });

  it("does not poll when there is no data (initial load or error)", () => {
    expect(poll(undefined)).toBe(false);
  });
});
