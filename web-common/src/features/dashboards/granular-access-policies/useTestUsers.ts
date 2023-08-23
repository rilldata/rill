import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export interface TestUser {
  name?: string;
  email?: string;
  groups?: string[];
  admin?: boolean;
}

export function useTestUsers(instanceId: string) {
  return createRuntimeServiceGetFile(instanceId, `rill.yaml`, {
    query: {
      select: (data) => {
        const yamlObj = parse(data?.blob);
        const testUsers = yamlObj?.test_users || [];
        return testUsers as Array<TestUser>;
      },
    },
  });
}
