export const IS_PRODUCTION = process.env.NODE_ENV === "production";

export const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";

export const ADMIN_HOST = IS_PRODUCTION
  ? "admin.rilldata.com"
  : "localhost:8080";
export const ADMIN_URL = `${HTTP_PROTOCOL}://${ADMIN_HOST}`;
