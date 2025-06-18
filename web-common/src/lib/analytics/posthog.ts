import posthog, { type Properties } from "posthog-js";

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;

export function initPosthog(rillVersion: string, sessionId?: string | null) {
  // No need to proceed if PostHog is already initialized
  if ((posthog as any).__loaded) return;

  if (!POSTHOG_API_KEY) {
    console.warn("PostHog API Key not found");
    return;
  }

  fetch("https://ipapi.co/json/")
    .then((res) => res.json())
    .then((data) => {
      const isUS = data && data.country_code === "US";
      const isCalifornia =
        isUS && (data.region_code === "CA" || data.region === "California");
      const shouldRedact = !isUS || isCalifornia;

      posthog.init(POSTHOG_API_KEY, {
        api_host: "https://us.i.posthog.com",
        session_recording: shouldRedact
          ? {
              maskAllInputs: true,
              maskTextSelector: "*",
              recordHeaders: true,
              recordBody: false,
            }
          : {}, // No redaction for non-California US users
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
    })
    .catch(() => {
      // Fallback: initialize with redaction if GeoIP fails
      posthog.init(POSTHOG_API_KEY, {
        api_host: "https://us.i.posthog.com",
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
    });
}

export function posthogIdentify(userID: string, userProperties?: Properties) {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access
  posthog.identify(userID, userProperties);
}

export function addPosthogSessionIdToUrl(url: string) {
  return url + "?ph_session_id=" + posthog.get_session_id();
}
