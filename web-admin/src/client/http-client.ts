import type { AxiosRequestConfig } from "axios";
import Axios from "axios";

/**
 * The canonical URL of the admin server.
 * It does not change when the frontend is running on a custom domain.
 */
export const CANONICAL_ADMIN_URL =
  import.meta.env.RILL_UI_PUBLIC_RILL_ADMIN_URL || "http://localhost:8080";

/**
 * The URL of the admin server.
 *
 * By convention, if the frontend detects that it is running on a custom domain,
 * instead of using the canonical admin URL, it should contact the admin server on /api on the same domain.
 *
 * The only exceptions to this rule are for /auth/login redirects (not other /auth endpoints) and for /github redirects,
 * which should always use the canonical admin URL.
 */
export const ADMIN_URL =
  typeof window === "undefined" ||
  urlExtractSLD(window.location.origin) === urlExtractSLD(CANONICAL_ADMIN_URL)
    ? CANONICAL_ADMIN_URL
    : urlRewritePath(window.location.origin, "/api");

/**
 * extractSLD extracts the second-level domain from the given URL.
 * For example, "www.example.com" returns "example.com" and "localhost:8080" returns "localhost".
 */
function urlExtractSLD(url: string): string {
  const parsed = new URL(url);
  const parts = parsed.hostname.split(".");
  if (parts.length <= 2) {
    return parts.join(".");
  }
  return parts.slice(-2).join(".");
}

/**
 * urlRewritePath rewrites the path of the given URL.
 */
function urlRewritePath(url: string, path: string): string {
  const parsed = new URL(url);
  parsed.pathname = path;
  return parsed.toString();
}

export const AXIOS_INSTANCE = Axios.create({
  baseURL: ADMIN_URL,
  withCredentials: true,
});

// TODO: use the new client?
export const httpClient = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const { data } = await AXIOS_INSTANCE(config);
  return data;
};

export default httpClient;
