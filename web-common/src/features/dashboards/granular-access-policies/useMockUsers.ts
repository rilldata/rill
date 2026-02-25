import { parseDocument } from "yaml";
import type { RuntimeClient } from "../../../runtime-client/v2";
import { createRuntimeServiceGetFile } from "../../../runtime-client/v2/gen/runtime-service";

export interface MockUser {
  email?: string;
  name?: string;
  admin?: boolean;
  groups?: string[];
  attributes?: { [key: string]: any };
}

export function useMockUsers(client: RuntimeClient) {
  return createRuntimeServiceGetFile(
    client,
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
        refetchOnMount: true,
      },
    },
  );
}
