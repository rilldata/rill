import type {
  V1GetDeploymentCredentialsResponse,
  V1GetProjectResponse,
  V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import type { AuthContext } from "@rilldata/web-common/runtime-client/v2/runtime-client";

/**
 * Resolves the effective runtime connection based on which auth mode is active.
 *
 * Three modes, in priority order:
 *   1. Mock (View As): use mocked user's credentials and permissions
 *   2. Magic (Public URL): bearer-token auth, project's runtime host/jwt
 *   3. User (default): cookie auth, project's runtime host/jwt
 */
export function resolveRuntimeConnection(
  projectData: V1GetProjectResponse | undefined,
  mockUser:
    | {
        credentials: V1GetDeploymentCredentialsResponse;
        permissions?: V1ProjectPermissions;
      }
    | undefined,
  onPublicURLPage: boolean,
): {
  authContext: AuthContext;
  host: string | undefined;
  instanceId: string | undefined;
  jwt: string | undefined;
  projectPermissions: V1ProjectPermissions | undefined;
} {
  if (mockUser) {
    return {
      authContext: "mock",
      host: mockUser.credentials.runtimeHost,
      instanceId: mockUser.credentials.instanceId,
      jwt: mockUser.credentials.accessToken,
      projectPermissions:
        mockUser.permissions ?? projectData?.projectPermissions,
    };
  }

  return {
    authContext: onPublicURLPage ? "magic" : "user",
    host: projectData?.deployment?.runtimeHost,
    instanceId: projectData?.deployment?.runtimeInstanceId,
    jwt: projectData?.jwt,
    projectPermissions: projectData?.projectPermissions,
  };
}
