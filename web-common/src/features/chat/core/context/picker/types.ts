import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { CreateQueryOptions } from "@tanstack/svelte-query";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client.ts";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";

export type InlineContextPickerOption = {
  context: InlineContext;
  recentlyUsed?: boolean;
  currentlyActive?: boolean;
  children?: InlineContext[][];
  childrenQueryOptions?: CreateQueryOptions<
    unknown,
    ErrorType<RpcStatus>,
    InlineContext[][]
  >;
};
