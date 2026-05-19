// The frontend runs in two distinct surfaces for editing files.
// Within a single page session the editEnvironment is fixed: cloud users never become local users, and vice versa.
// The flag is set once at the surface's entry point and read by code that needs to
// vary behavior per editEnvironment (e.g. `.env` is readonly on cloud but editable locally).

import { updateLocalFileSchemaForCloud } from "@rilldata/web-common/features/templates/schemas/local_file.ts";

export type RuntimeEditEnvironment = "local" | "cloud";

let editEnvironment: RuntimeEditEnvironment = "local";

export function setRuntimeEditEnvironment(env: RuntimeEditEnvironment) {
  editEnvironment = env;
  if (env === "cloud") {
    updateLocalFileSchemaForCloud();
  }
}

export function isCloudRuntimeEditEnvironment() {
  return editEnvironment === "cloud";
}
