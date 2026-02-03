import posthog, { type Properties } from "posthog-js";

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;

export function initPosthog(rillVersion: string, sessionId?: string | null) {
  console.log("[PostHog] initPosthog called, API key exists:", !!POSTHOG_API_KEY);

  // No need to proceed if PostHog is already initialized
  if (posthog.__loaded) {
    console.log("[PostHog] Already loaded, skipping init");
    return;
  }

  if (!POSTHOG_API_KEY) {
    console.warn("PostHog API Key not found");
    return;
  }

  // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access
  posthog.init(POSTHOG_API_KEY, {
    api_host: "https://us.i.posthog.com", // TODO: use a reverse proxy https://posthog.com/docs/advanced/proxy
    session_recording: {
      // Selective input masking by type
      maskAllInputs: false,
      maskInputOptions: {
        password: true, // Always mask passwords
        email: true, // Mask emails (customer data)
        tel: true, // Mask phone numbers (customer data)
        // Don't mask these types
        color: false,
        date: false,
        number: false,
        search: false,
        text: false,
        url: false,
      },
      // Custom masking function for granular control
      maskInputFn: (text, element) => {
        // Always mask passwords
        if (element?.type === "password") {
          return "*".repeat(text.length);
        }
        // Mask inputs marked as sensitive via data attribute
        if (element?.getAttribute("data-sensitive") === "true") {
          return "*".repeat(text.length);
        }
        // Mask email and phone inputs
        if (element?.type === "email" || element?.type === "tel") {
          return "*".repeat(text.length);
        }
        // Show everything else (UI elements remain visible)
        return text;
      },
      // Keep UI text visible (buttons, labels, navigation)
      maskTextSelector: undefined,
      // Network request settings
      recordHeaders: true,
      recordBody: false,
    },
    autocapture: true,
    enable_heatmaps: true,
    disable_session_recording: false, // Explicitly enable session recording
    bootstrap: {
      sessionID: sessionId ?? undefined,
    },
    loaded: (posthog) => {
      posthog.register_for_session({
        "Rill version": rillVersion,
      });
      console.log("[PostHog] Initialized, session recording:", posthog.sessionRecordingStarted());
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
