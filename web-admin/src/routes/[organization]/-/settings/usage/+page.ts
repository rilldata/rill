import {
  adminServiceGetEmbeddedAnalytics,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { error } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  try {
    const resp = await adminServiceGetEmbeddedAnalytics(params.organization);
    return {
      iframeUrl: resp.iframeUrl ?? "",
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Failed to fetch embedded analytics");
    }

    throw error(e.response.status, e.response.data.message);
  }
};
