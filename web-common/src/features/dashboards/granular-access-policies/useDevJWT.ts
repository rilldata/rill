import { createRuntimeServiceIssueDevJWT } from "../../../runtime-client";
import type { MockUser } from "./useMockUsers";

export function useDevJWT(mockUser: MockUser | null) {
  return createRuntimeServiceIssueDevJWT(
    {
      name: mockUser?.name ? mockUser.name : "Mock User",
      email: mockUser?.email,
      groups: mockUser?.groups ? mockUser.groups : [],
      admin: mockUser?.admin ? true : false,
    },
    {
      query: {
        enabled: mockUser !== null,
        // TODO: this should move to a global error handler at the QueryCache level
        onError: (err) => {
          console.error("Error issuing Dev JWT", err);
        },
      },
    }
  );
}
