import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { CanvasValidResponse } from "@rilldata/web-common/features/canvas/selector";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";

export type Vector = [number, number];

export type CanvasSpecResponseStore = Readable<
  QueryObserverResult<CanvasValidResponse, RpcStatus>
>;
