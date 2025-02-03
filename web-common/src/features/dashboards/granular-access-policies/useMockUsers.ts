import { parseDocument } from "yaml";
import { createRuntimeServiceGetFile } from "../../../runtime-client";

export interface MockUser {
  email?: string;
  name?: string;
  admin?: boolean;
  groups?: string[];
  attributes?: { [key: string]: any };
}

export function useMockUsers(instanceId: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: `rill.yaml` },
    {
      query: {
        select: (data) => {
          if (!data.blob) return [];
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
