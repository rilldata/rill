import { getSafeRedirectTarget } from "@rilldata/web-admin/lib/safe-redirect.ts";

export type DeviceAuthorizationInput = {
  userCode: string | null;
  confirmed: boolean;
  redirect: string | null;
  currentUrl: string;
};

export type DeviceAuthorizationState = {
  inFlight: boolean;
  completed: boolean;
  successMessage: string;
  errorMessage: string;
};

export type DeviceAuthorizationMessages = {
  confirmed: string;
  rejected: string;
  confirmationFailed: string;
  rejectionFailed: string;
  invalidCode: string;
  unsafeRedirect: string;
};

type DeviceAuthorizationResponse = {
  ok: boolean;
  text: () => Promise<string>;
};

export type DeviceAuthorizationDependencies = {
  adminUrl: string;
  fetch: (
    input: string,
    init: RequestInit,
  ) => Promise<DeviceAuthorizationResponse>;
  navigate: (url: string) => void;
  messages: DeviceAuthorizationMessages;
  onState: (state: DeviceAuthorizationState) => void;
};

export function isValidDeviceUserCode(code: string | null): code is string {
  return !!code && code.length <= 128 && /^[A-Za-z0-9_-]+$/.test(code);
}

export function buildDeviceAuthorizationRequestUrl(
  adminUrl: string,
  userCode: string,
  confirmed: boolean,
): string {
  const requestUrl = new URL(adminUrl);
  const basePath = requestUrl.pathname.replace(/\/+$/, "");
  requestUrl.pathname = `${basePath}/auth/oauth/device`;
  requestUrl.search = "";
  requestUrl.hash = "";

  // URLSearchParams owns encoding, so user input can never add or override a
  // request parameter through string concatenation.
  requestUrl.searchParams.set("user_code", userCode);
  requestUrl.searchParams.set("code_confirmed", String(confirmed));
  return requestUrl.toString();
}

function withDetail(message: string, detail: string) {
  const trimmedDetail = detail.trim();
  return trimmedDetail ? `${message}: ${trimmedDetail}` : message;
}

async function performDeviceAuthorization(
  input: DeviceAuthorizationInput,
  dependencies: DeviceAuthorizationDependencies,
): Promise<Pick<DeviceAuthorizationState, "successMessage" | "errorMessage">> {
  const failureMessage = input.confirmed
    ? dependencies.messages.confirmationFailed
    : dependencies.messages.rejectionFailed;

  if (!isValidDeviceUserCode(input.userCode)) {
    throw new Error(dependencies.messages.invalidCode);
  }

  let response: DeviceAuthorizationResponse;
  try {
    response = await dependencies.fetch(
      buildDeviceAuthorizationRequestUrl(
        dependencies.adminUrl,
        input.userCode,
        input.confirmed,
      ),
      {
        method: "POST",
        credentials: "include",
      },
    );
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(withDetail(failureMessage, detail));
  }

  if (!response.ok) {
    let detail = "";
    try {
      detail = await response.text();
    } catch {
      // The status still provides a useful failure when the body is unreadable.
    }
    throw new Error(withDetail(failureMessage, detail));
  }

  if (!input.confirmed) {
    return {
      successMessage: dependencies.messages.rejected,
      errorMessage: "",
    };
  }

  if (input.redirect) {
    const safeRedirect = getSafeRedirectTarget(
      input.redirect,
      input.currentUrl,
      { allowLoopback: true },
    );
    if (safeRedirect) {
      dependencies.navigate(safeRedirect);
    } else {
      return {
        successMessage: dependencies.messages.confirmed,
        errorMessage: dependencies.messages.unsafeRedirect,
      };
    }
  }

  return {
    successMessage: dependencies.messages.confirmed,
    errorMessage: "",
  };
}

export function createDeviceAuthorizationAction(
  dependencies: DeviceAuthorizationDependencies,
) {
  let inFlight: Promise<void> | null = null;
  let completed = false;

  return function authorizeDevice(
    input: DeviceAuthorizationInput,
  ): Promise<void> {
    // A pending or completed decision owns the device code. Repeated clicks
    // must not send a second confirm/reject request.
    if (inFlight !== null) return inFlight;
    if (completed) return Promise.resolve();

    dependencies.onState({
      inFlight: true,
      completed: false,
      successMessage: "",
      errorMessage: "",
    });

    const request = performDeviceAuthorization(input, dependencies)
      .then(({ successMessage, errorMessage }) => {
        completed = true;
        dependencies.onState({
          inFlight: false,
          completed: true,
          successMessage,
          errorMessage,
        });
      })
      .catch((error) => {
        dependencies.onState({
          inFlight: false,
          completed: false,
          successMessage: "",
          errorMessage: error instanceof Error ? error.message : String(error),
        });
      })
      .finally(() => {
        // Request failures unlock the buttons for a deliberate retry; success
        // remains completed and cannot be submitted again.
        if (inFlight === request) inFlight = null;
      });

    inFlight = request;
    return request;
  };
}
