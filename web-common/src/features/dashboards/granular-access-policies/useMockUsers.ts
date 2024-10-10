import { parseDocument } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export interface MockUser {
  name?: string;
  email?: string;
  groups?: string[];
  admin?: boolean;
}

export function useMockUsers(instanceId: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: `rill.yaml` },
    {
      query: {
        select: (data) => {
          const yamlObj = parseDocument(data.blob, {
            logLevel: "error",
          })?.toJS();
          const mockUsers =
            yamlObj?.mock_users?.filter((user: MockUser) => user?.email) || [];
          return mockUsers as Array<MockUser>;
        },
      },
    },
  );
}
