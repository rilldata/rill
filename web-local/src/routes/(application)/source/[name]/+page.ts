import { runtimeServiceGetCatalogObject } from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import { fetchWrapper } from "../../../../lib/util/fetchWrapper";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");
    const sourceResp = await runtimeServiceGetCatalogObject(
      instanceResp.instanceId,
      params.name
    );

    return {
      sourceName: sourceResp.object.source.name,
    };
  } catch (e) {
    if (e.response?.status && e.response?.data?.message) {
      throw error(e.response.status, e.response.data.message);
    } else {
      console.error(e);
      throw error(500, e.message);
    }
  }
}
