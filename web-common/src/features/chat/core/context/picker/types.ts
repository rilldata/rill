import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { CreateQueryOptions } from "@tanstack/svelte-query";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client.ts";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable, Writable } from "svelte/store";

// Sections are aggregated group of parent options by type like all metrics views, all models, etc.
export type InlineContextPickerSection = InlineContextPickerParentOption[];

export type InlineContextPickerParentOption = {
  context: InlineContext;
  openStore: Writable<boolean>;
  recentlyUsed?: boolean;
  currentlyActive?: boolean;
  children?: InlineContext[][];
  childrenQueryOptions?: Readable<
    CreateQueryOptions<unknown, ErrorType<RpcStatus>, InlineContext[][]>
  >;
  childrenLoading?: boolean;
};
