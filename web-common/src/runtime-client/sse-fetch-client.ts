/**
 * @deprecated Import from `@rilldata/web-common/runtime-client/sse` instead.
 * This shim keeps existing imports compiling; it will be removed once all
 * consumers have migrated to the new layered API.
 */
export {
  SSEFetchClient,
  SSEHttpError,
  type SSEMessage,
} from "./sse/sse-fetch-client";
