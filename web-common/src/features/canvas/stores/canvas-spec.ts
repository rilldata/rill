import type { QueryObserverResult } from "@rilldata/svelte-query";
import { type CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type {
  RpcStatus,
  V1CanvasSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

export class CanvasResolvedSpec {
  canvasSpec: Readable<V1CanvasSpec | undefined>;
  isLoading: Readable<boolean>;

  constructor(
    validSpecStore: Readable<QueryObserverResult<CanvasResponse, RpcStatus>>,
  ) {
    this.canvasSpec = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.data?.canvas;
    });

    this.isLoading = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.isLoading;
    });
  }
}
