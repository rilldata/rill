/**
 * Hack: The `GetProject` API returns the host with the wrong port.
 * In development, the runtime host is actually on port 8081.
 */
export function fixLocalhostRuntimePort(host: string) {
  return host.replace("localhost:9091", "localhost:8081");
}
