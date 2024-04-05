import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";
import { get } from "svelte/store";
import { RuntimeUrl } from "../lib/application-state-stores/initialize-node-store-contexts.js";

export const ssr = false;

export async function load(opts) {
  console.log(opts);
  console.log(RILL_RUNTIME_URL);
  //   console.log(localStorage.getItem("token"));

  return {
    props: {
      // we can pass some initial props to the app here
    },
  };
}
