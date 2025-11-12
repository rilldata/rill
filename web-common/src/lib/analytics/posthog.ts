import posthog, {
  type Properties,
  type CaptureResult,
  type BeforeSendFn,
} from "posthog-js";

const allowedExceptionHostnames = new Set([
  "ui.rilldata.com",
  "localhost",
  "127.0.0.1",
]);

const filterExceptionEvents: BeforeSendFn = (event: CaptureResult | null) => {
  if (event?.event !== "$exception") {
    return event;
  }

  if (typeof window === "undefined") {
    return null;
  }

  // For $exception events, only send them from specific hostnames.
  // This was motivated by the desire to decrease our potential for large event tracking bills from PostHog.
  const hostname = window.location.hostname;
  if (allowedExceptionHostnames.has(hostname)) {
    return event;
  }

  return null;
};

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;

export function initPosthog(rillVersion: string, sessionId?: string | null) {
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
    before_send: filterExceptionEvents,
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
