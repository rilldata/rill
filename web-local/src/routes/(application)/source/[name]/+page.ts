import { runtimeServiceGetCatalogObject } from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import { LOCAL_RUNTIME_INSTANCE_ID } from "../../../../lib/config/constants";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  // TODO: might have to check for if(browser) here

  try {
    await runtimeServiceGetCatalogObject(
      LOCAL_RUNTIME_INSTANCE_ID,
      params.name
    );

    console.log("source name", params.name);
    return {
      runtimeInstanceId: LOCAL_RUNTIME_INSTANCE_ID,
      sourceName: params.name,
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
