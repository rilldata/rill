import {
  adminServiceGetProjectEmbeddedAnalytics,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { error } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  try {
    const resp = await adminServiceGetProjectEmbeddedAnalytics(
      params.organization,
      params.project,
    );
    return {
      iframeUrl: resp.iframeUrl ?? "",
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Failed to fetch project analytics");
    }

    throw error(e.response.status, e.response.data.message);
  }
};
