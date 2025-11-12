import posthog, {
  type Properties,
  type CaptureResult,
  type BeforeSendFn,
} from "posthog-js";

type ExceptionFilterConfig = {
  allowedHostnames?: string[];
  allowedDomainSuffixes?: string[];
  /**
   * Default sampling rate used for hosts that are not explicitly allowed.
   * 1.0 = send all events, 0 = drop all events.
   */
  defaultSampleRate?: number;
  /**
   * Per-host sampling overrides. Hostnames should be provided in lowercase.
   */
  hostSampleRates?: Record<string, number>;
};

declare global {
  interface Window {
    __RILL_POSTHOG_EXCEPTION_FILTER__?: ExceptionFilterConfig;
  }
}

const DEFAULT_ALLOWED_HOSTNAMES = ["ui.rilldata.com", "localhost", "127.0.0.1"];
const DEFAULT_ALLOWED_DOMAIN_SUFFIXES = [".rilldata.com"];
const DEFAULT_SAMPLE_RATE = 0.1;

const normalizeHost = (host: string) => host.trim().toLowerCase();

const resolveExceptionFilterOptions = () => {
  if (typeof window === "undefined") {
    return {
      allowedHostnames: new Set<string>(DEFAULT_ALLOWED_HOSTNAMES.map(normalizeHost)),
      allowedDomainSuffixes: DEFAULT_ALLOWED_DOMAIN_SUFFIXES.map(normalizeHost),
      hostSampleRates: {} as Record<string, number>,
      defaultSampleRate: DEFAULT_SAMPLE_RATE,
    };
  }

  const runtimeConfig = window.__RILL_POSTHOG_EXCEPTION_FILTER__ ?? {};
  const allowedHostnames = new Set(
    [...DEFAULT_ALLOWED_HOSTNAMES, ...(runtimeConfig.allowedHostnames ?? [])].map(
      normalizeHost,
    ),
  );
  const allowedDomainSuffixes = [
    ...DEFAULT_ALLOWED_DOMAIN_SUFFIXES,
    ...(runtimeConfig.allowedDomainSuffixes ?? []),
  ].map((suffix) => {
    const normalized = normalizeHost(suffix);
    return normalized.startsWith(".") ? normalized.slice(1) : normalized;
  });
  const hostSampleRates = Object.fromEntries(
    Object.entries(runtimeConfig.hostSampleRates ?? {}).map(([key, value]) => [
      normalizeHost(key),
      value,
    ]),
  );
  const defaultSampleRate =
    typeof runtimeConfig.defaultSampleRate === "number"
      ? runtimeConfig.defaultSampleRate
      : DEFAULT_SAMPLE_RATE;

  return {
    allowedHostnames,
    allowedDomainSuffixes,
    hostSampleRates,
    defaultSampleRate,
  };
};

const isAllowedHostname = (
  hostname: string,
  allowedHostnames: Set<string>,
  allowedDomainSuffixes: string[],
) => {
  if (allowedHostnames.has(hostname)) return true;

  return allowedDomainSuffixes.some((suffix) => {
    if (!suffix.length) return false;
    return hostname === suffix || hostname.endsWith(`.${suffix}`);
  });
};

const clampSampleRate = (rate: number) => {
  if (Number.isNaN(rate)) return 0;
  if (rate < 0) return 0;
  if (rate > 1) return 1;
  return rate;
};

const filterExceptionEvents: BeforeSendFn = (event: CaptureResult | null) => {
  if (event?.event !== "$exception") {
    return event;
  }

  if (typeof window === "undefined") {
    return event;
  }

  const { allowedHostnames, allowedDomainSuffixes, hostSampleRates, defaultSampleRate } =
    resolveExceptionFilterOptions();

  const hostname = normalizeHost(window.location.hostname);

  // Always preserve exceptions for allowed hosts and embeds.
  if (
    isAllowedHostname(hostname, allowedHostnames, allowedDomainSuffixes) ||
    window.location.pathname.includes("/-/embed/")
  ) {
    return event;
  }

  const sampleRate = clampSampleRate(
    hostSampleRates[hostname] ?? defaultSampleRate,
  );
  if (sampleRate >= 1) {
    return event;
  }
  if (sampleRate <= 0) {
    return null;
  }

  return Math.random() < sampleRate ? event : null;
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
