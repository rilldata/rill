import type {
  V1GetDeploymentCredentialsResponse,
  V1GetProjectResponse,
  V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { describe, expect, it } from "vitest";
import { resolveRuntimeConnection } from "./project-runtime";

const fakeProjectData: V1GetProjectResponse = {
  deployment: {
    runtimeHost: "https://runtime.example.com",
    runtimeInstanceId: "inst-123",
  },
  jwt: "project-jwt",
  projectPermissions: {
    readDev: true,
    manageDev: false,
    manageProjectMembers: true,
    createMagicAuthTokens: true,
  },
};

const fakeMockedCredentials: V1GetDeploymentCredentialsResponse = {
  runtimeHost: "https://mock-runtime.example.com",
  instanceId: "mock-inst-456",
  accessToken: "mock-jwt",
};

const fakeMockedPermissions: V1ProjectPermissions = {
  readDev: false,
  manageDev: false,
  manageProjectMembers: false,
  createMagicAuthTokens: false,
};

describe("resolveRuntimeConnection", () => {
  it("returns cookie auth (user) by default", () => {
    const result = resolveRuntimeConnection(fakeProjectData, undefined, false);
    expect(result).toEqual({
      authContext: "user",
      host: "https://runtime.example.com",
      instanceId: "inst-123",
      jwt: "project-jwt",
      projectPermissions: fakeProjectData.projectPermissions,
    });
  });

  it("returns undefined fields when projectData is undefined", () => {
    const result = resolveRuntimeConnection(undefined, undefined, false);
    expect(result).toEqual({
      authContext: "user",
      host: undefined,
      instanceId: undefined,
      jwt: undefined,
      projectPermissions: undefined,
    });
  });

  it("returns magic auth on public URL pages", () => {
    const result = resolveRuntimeConnection(fakeProjectData, undefined, true);
    expect(result.authContext).toBe("magic");
    expect(result.host).toBe("https://runtime.example.com");
  });

  it("returns mock auth when View As is active", () => {
    const result = resolveRuntimeConnection(
      fakeProjectData,
      {
        credentials: fakeMockedCredentials,
        permissions: fakeMockedPermissions,
      },
      false,
    );
    expect(result).toEqual({
      authContext: "mock",
      host: "https://mock-runtime.example.com",
      instanceId: "mock-inst-456",
      jwt: "mock-jwt",
      projectPermissions: fakeMockedPermissions,
    });
  });

  it("mock auth takes priority over public URL", () => {
    const result = resolveRuntimeConnection(
      fakeProjectData,
      {
        credentials: fakeMockedCredentials,
        permissions: fakeMockedPermissions,
      },
      true,
    );
    expect(result.authContext).toBe("mock");
  });

  it("falls back to project permissions when mock permissions are undefined", () => {
    const result = resolveRuntimeConnection(
      fakeProjectData,
      { credentials: fakeMockedCredentials },
      false,
    );
    expect(result.projectPermissions).toEqual(
      fakeProjectData.projectPermissions,
    );
  });
});
