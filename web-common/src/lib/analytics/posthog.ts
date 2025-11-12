import posthog, { type Properties } from "posthog-js";

let exceptionCaptureFilterRegistered = false;

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;

export function initPosthog(rillVersion: string, sessionId?: string | null) {
  if (!exceptionCaptureFilterRegistered && typeof window !== "undefined") {
    exceptionCaptureFilterRegistered = true;
    posthog.on("capture", (event) => {
      if (event?.event !== "$exception") {
        return event;
      }

      // For $exception events, only send them from specific hostnames.
      // This was motivated by the desire to decrease our potential for large event tracking bills from PostHog.
      const hostname = window.location.hostname;
      const allowedHostnames = ["ui.rilldata.com", "localhost", "127.0.0.1"];
      if (allowedHostnames.includes(hostname)) {
        return event;
      }

      // Drop $exception events for all other hostnames.
      return null;
    });
  }

  // No need to proceed if PostHog is already initialized
  if (posthog.__loaded) return;

  if (!POSTHOG_API_KEY) {
    console.warn("PostHog API Key not found");
    return;
  }

  // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access
  posthog.init(POSTHOG_API_KEY, {
    api_host: "https://us.i.posthog.com", // TODO: use a reverse proxy https://posthog.com/docs/advanced/proxy
    session_recording: {
      maskAllInputs: true,
      maskTextSelector: "*",
      recordHeaders: true,
      recordBody: false,
    },
    autocapture: true,
    enable_heatmaps: true,
    bootstrap: {
      sessionID: sessionId ?? undefined,
    },
    loaded: (posthog) => {
      posthog.register_for_session({
        "Rill version": rillVersion,
      });
    },
  });
}

export function posthogIdentify(userID: string, userProperties?: Properties) {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access
  posthog.identify(userID, userProperties);
}

export function addPosthogSessionIdToUrl(url: string) {
  const u = new URL(url);
  u.searchParams.set("ph_session_id", posthog.get_session_id());
  return u.toString();
}
