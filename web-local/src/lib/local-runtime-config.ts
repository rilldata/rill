// Detect dev mode without depending on SvelteKit's $app/environment
// (which is unavailable to tsc outside Vite/SvelteKit context)
const isDev =
  typeof import.meta.env !== "undefined" && import.meta.env.DEV === true;

export const LOCAL_HOST = isDev ? "http://localhost:9009" : "";
export const LOCAL_INSTANCE_ID = "default";
