import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage.ts";

export function getCanvasModeStore(
  organization: string,
  project: string,
  name: string,
) {
  return sessionStorageStore(
    `app:rill:${organization}:${project}:${name}`,
    "view",
  );
}

export function setCanvasMode(
  organization: string,
  project: string,
  name: string,
  mode: "view" | "edit",
) {
  sessionStorage.setItem(
    `app:rill:${organization}:${project}:${name}`,
    JSON.stringify(mode),
  );
}
