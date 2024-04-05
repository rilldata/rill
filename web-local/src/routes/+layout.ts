// When testing, we need to use the relative path to the server
const host = import.meta.env.DEV ? "http://localhost:9009" : "";

export const ssr = false;

export function load() {
  return {
    host,
    instanceId: "default",
  };
}
