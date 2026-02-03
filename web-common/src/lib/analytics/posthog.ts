import posthog, { type Properties } from "posthog-js";

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;
const CONSENT_KEY = "rill_session_recording_consent";

export type SessionRecordingConsent = "granted" | "denied" | null;

/**
 * Get the stored session recording consent preference
 */
export function getSessionRecordingConsent(): SessionRecordingConsent {
  if (typeof window === "undefined") return null;
  const value = localStorage.getItem(CONSENT_KEY);
  if (value === "granted" || value === "denied") return value;
  return null;
}

/**
 * Store the session recording consent preference and apply it
 */
export function setSessionRecordingConsent(consent: "granted" | "denied") {
  if (typeof window === "undefined") return;
  localStorage.setItem(CONSENT_KEY, consent);
  applySessionRecordingConsent(consent);
}

/**
 * Apply session recording settings based on consent
 */
function applySessionRecordingConsent(consent: SessionRecordingConsent) {
  if (!posthog.__loaded) return;

  if (consent === "granted") {
    // Start session recording with selective masking (UI visible)
    posthog.startSessionRecording();
  } else if (consent === "denied") {
    // Start session recording with full masking (privacy mode)
    // PostHog doesn't support changing mask config at runtime,
    // so we stop recording entirely for declined consent
    posthog.stopSessionRecording();
  }
}

export function initPosthog(rillVersion: string, sessionId?: string | null) {
  // No need to proceed if PostHog is already initialized
  if (posthog.__loaded) return;

  if (!POSTHOG_API_KEY) {
    console.warn("PostHog API Key not found");
    return;
  }

  const consent = getSessionRecordingConsent();

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
        const inputElement = element as HTMLInputElement | undefined;
        // Always mask passwords
        if (inputElement?.type === "password") {
          return "*".repeat(text.length);
        }
        // Mask inputs marked as sensitive via data attribute
        if (element?.getAttribute("data-sensitive") === "true") {
          return "*".repeat(text.length);
        }
        // Mask email and phone inputs
        if (inputElement?.type === "email" || inputElement?.type === "tel") {
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
    // Start with session recording disabled until consent is given
    disable_session_recording: consent !== "granted",
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
