import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";

export type CanvasSpecResponseStore = Readable<
  QueryObserverResult<CanvasResponse, RpcStatus>
>;
