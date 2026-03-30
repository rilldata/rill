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
  mockedUserId: string | undefined,
  mockedCredentials: V1GetDeploymentCredentialsResponse | undefined,
  mockedProjectPermissions: V1ProjectPermissions | undefined,
  onPublicURLPage: boolean,
): {
  authContext: AuthContext;
  host: string | undefined;
  instanceId: string | undefined;
  jwt: string | undefined;
  projectPermissions: V1ProjectPermissions | undefined;
} {
  const isMocked = !!(mockedUserId && mockedCredentials);

  if (isMocked) {
    return {
      authContext: "mock",
      host: mockedCredentials.runtimeHost,
      instanceId: mockedCredentials.instanceId,
      jwt: mockedCredentials.accessToken,
      projectPermissions:
        mockedProjectPermissions ?? projectData?.projectPermissions,
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
