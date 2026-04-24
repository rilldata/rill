/**
 * @deprecated Import from `@rilldata/web-common/runtime-client/sse` instead.
 * This shim keeps existing imports compiling; it will be removed once all
 * consumers have migrated to the new layered API.
 *
 * The class was renamed `SSEConnectionManager` → `SSEConnection`. The old
 * name is re-exported here as an alias.
 */
export {
  SSEConnection as SSEConnectionManager,
  ConnectionStatus,
} from "./sse/sse-connection";
