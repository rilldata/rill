import * as m from "@rilldata/web-common/paraglide/messages.js";
import { type UserFacingError } from "./error-store";

export function createUserFacingError(
  status: number | null,
  message: string,
): UserFacingError {
  // Handle network errors
  if (message === "Network Error") {
    return {
      statusCode: null,
      header: m.error_network_header(),
      body: m.error_network_body(),
    };
  }

  // Handle some application errors
  if (status === 400 && message === "driver: not found") {
    return {
      statusCode: status,
      header: m.error_deployment_not_found_header(),
      body: m.error_deployment_not_found_body(),
    };
  } else if (status === 401 && message === "auth token is expired") {
    return {
      statusCode: 401,
      header: m.error_link_expired_header(),
      body: m.error_link_expired_body(),
      fatal: true,
    };
  } else if (status === 401) {
    return {
      statusCode: 401,
      header: m.error_auth_header(),
      body: m.error_auth_body(),
    };
  } else if (status === 403) {
    return {
      statusCode: status,
      header: m.error_access_denied_header(),
      body: m.error_access_denied_body(),
    };
  } else if (message === "org not found") {
    return {
      statusCode: status,
      header: m.error_org_not_found_header(),
      body: m.error_org_not_found_body(),
    };
  } else if (message === "project not found") {
    return {
      statusCode: status,
      header: m.error_project_not_found_header(),
      body: m.error_project_not_found_body(),
    };
  } else if (
    status === 400 &&
    message.includes("failed to find the conversation")
  ) {
    return {
      statusCode: 404,
      header: m.error_conversation_not_found_header(),
      body: m.error_conversation_not_found_body(),
    };
  } else if (status === 404 && message === "resource not found") {
    return {
      statusCode: 404,
      header: m.error_resource_not_found_header(),
      body: m.error_resource_not_found_body(),
    };
  } else if (status === 404) {
    return {
      statusCode: 404,
      header: m.error_page_not_found_header(),
      body: m.error_page_not_found_body(),
    };
  }

  // Fallback for all other errors (including 5xx errors)
  return {
    statusCode: status,
    header: m.error_generic_header(),
    body: m.error_generic_body(),
    detail: message,
  };
}
