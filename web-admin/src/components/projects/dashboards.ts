import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import Axios from "axios";

export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
) {
  // Hack: in development, the runtime host is actually on port 8081
  const runtimeHost = projectData.productionDeployment.runtimeHost.replace(
    "localhost:9091",
    "localhost:8081"
  );

  const axios = Axios.create({
    baseURL: runtimeHost,
    headers: {
      Authorization: `Bearer ${projectData.jwt}`,
    },
  });

  const { data } = await axios.get(
    `/v1/instances/${projectData.productionDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
  );

  return data.entries;
}
