// This files contains clients that are not written through GRPC

import httpClient from "@rilldata/web-common/runtime-client/http-client";

export const runtimeServiceFileUpload = async (
  instanceId: string,
  filePath: string,
  formData: FormData,
) => {
  return httpClient({
    url: `/v1/instances/${instanceId}/files/upload/-/${filePath}`,
    method: "POST",
    data: formData,
    headers: {},
  });
};
