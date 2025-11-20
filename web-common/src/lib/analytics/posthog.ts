import posthog, { type Properties } from "posthog-js";

const POSTHOG_API_KEY = import.meta.env.RILL_UI_PUBLIC_POSTHOG_API_KEY;
const MASK_CHARACTER = "*";
const REGULATED_REGION_PROPERTY = "is_regulated_region";
const GEOIP_COUNTRY_PROPERTY = "$geoip_country_code";
const GEOIP_SUBDIVISION_PROPERTY = "$geoip_subdivision_1_code";

const EU_COUNTRIES = new Set([
  "AT",
  "BE",
  "BG",
  "HR",
  "CY",
  "CZ",
  "DK",
  "EE",
  "FI",
  "FR",
  "DE",
  "GR",
  "HU",
  "IE",
  "IT",
  "LV",
  "LT",
  "LU",
  "MT",
  "NL",
  "PL",
  "PT",
  "RO",
  "SK",
  "SI",
  "ES",
  "SE",
]);

type PosthogWithGeo = typeof posthog & {
  get_property?: (property: string) => unknown;
  set_person_properties?: (properties: Record<string, unknown>) => void;
  onFeatureFlags?: (callback: () => void) => void;
};

const maskValue = (text?: string | null) =>
  text ? MASK_CHARACTER.repeat(text.length) : (text ?? "");

export function initPosthog(rillVersion: string, sessionId?: string | null) {
  // No need to proceed if PostHog is already initialized
  if (posthog.__loaded) return;

  if (!POSTHOG_API_KEY) {
    console.warn("PostHog API Key not found");
    return;
  }

  const geoAwarePosthog = posthog as PosthogWithGeo;
  let isRegulatedUser = true;
  let persistedRegulatedStatus: boolean | undefined;

  const persistRegulatedStatus = (value: boolean) => {
    if (persistedRegulatedStatus === value) return;
    geoAwarePosthog.set_person_properties?.({
      [REGULATED_REGION_PROPERTY]: value,
    });
    persistedRegulatedStatus = value;
  };

  const evaluateGeoRegulation = () => {
    const existingValue = geoAwarePosthog.get_property?.(
      REGULATED_REGION_PROPERTY,
    ) as unknown;
    if (typeof existingValue === "boolean") {
      isRegulatedUser = existingValue;
      persistRegulatedStatus(existingValue);
      return;
    }
    if (typeof existingValue === "string") {
      const normalized = existingValue.toLowerCase();
      if (normalized === "true" || normalized === "false") {
        const boolValue = normalized === "true";
        isRegulatedUser = boolValue;
        persistRegulatedStatus(boolValue);
        return;
      }
    }

    const countryRaw = geoAwarePosthog.get_property?.(
      GEOIP_COUNTRY_PROPERTY,
    ) as unknown;
    const subdivisionRaw = geoAwarePosthog.get_property?.(
      GEOIP_SUBDIVISION_PROPERTY,
    ) as unknown;

    const normalizeRegionCode = (value: unknown) => {
      if (typeof value === "string") return value.toUpperCase();
      if (typeof value === "number") return value.toString().toUpperCase();
      return undefined;
    };

    const country = normalizeRegionCode(countryRaw);
    const subdivision = normalizeRegionCode(subdivisionRaw);

    if (!country) return;

    const computed =
      EU_COUNTRIES.has(country) || (country === "US" && subdivision === "CA");
    isRegulatedUser = computed;
    persistRegulatedStatus(computed);
  };

  const geoMaskInputFn = (
    text: string,
    element?: HTMLInputElement | HTMLTextAreaElement | null,
  ) => {
    if (!text) return text;
    const inputType = element?.getAttribute?.("type")?.toLowerCase();
    if (inputType === "password") {
      return maskValue(text);
    }
    if (!isRegulatedUser) return text;
    return maskValue(text);
  };

  posthog.init(POSTHOG_API_KEY, {
    api_host: "https://us.i.posthog.com", // TODO: use a reverse proxy https://posthog.com/docs/advanced/proxy
    session_recording: {
      maskAllInputs: false,
      maskAllText: false,
      maskInputFn: geoMaskInputFn,
      maskInputOptions: {
        password: true,
      },
      recordHeaders: true,
      recordBody: false,
    },
    autocapture: true,
    enable_heatmaps: true,
    bootstrap: {
      sessionID: sessionId ?? undefined,
    },
    loaded: (client) => {
      client.register_for_session({
        "Rill version": rillVersion,
      });
      evaluateGeoRegulation();
      geoAwarePosthog.onFeatureFlags?.(evaluateGeoRegulation);
    },
  });
}

export function posthogIdentify(userID: string, userProperties?: Properties) {
  posthog.identify(userID, userProperties);
}

export function addPosthogSessionIdToUrl(url: string) {
  const u = new URL(url);
  u.searchParams.set("ph_session_id", posthog.get_session_id());
  return u.toString();
}
