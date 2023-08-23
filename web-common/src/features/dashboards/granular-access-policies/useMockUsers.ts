import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export interface MockUser {
  name?: string;
  email?: string;
  groups?: string[];
  admin?: boolean;
}

export function useMockUsers(instanceId: string) {
  return createRuntimeServiceGetFile(instanceId, `rill.yaml`, {
    query: {
      select: (data) => {
        const yamlObj = parse(data?.blob);
        const mockUsers = yamlObj?.mock_users || [];
        return mockUsers as Array<MockUser>;
      },
    },
  });
}
