<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../components/button/Button.svelte";
  import {
    getSessionRecordingConsent,
    setSessionRecordingConsent,
  } from "./posthog";

  let showBanner = false;

  // EU country codes (ISO 3166-1 alpha-2)
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
    // EEA countries
    "IS",
    "LI",
    "NO",
    // UK (still follows similar privacy laws)
    "GB",
  ]);

  async function isEUOrCalifornia(): Promise<boolean> {
    try {
      const response = await fetch("https://ipapi.co/json/");
      if (!response.ok) return true; // Default to showing banner if geo fails
      const data = await response.json();

      const isEU = EU_COUNTRIES.has(data.country_code);
      const isCalifornia =
        data.country_code === "US" && data.region_code === "CA";

      return isEU || isCalifornia;
    } catch {
      // If geolocation fails, default to showing banner (safer for compliance)
      return true;
    }
  }

  onMount(async () => {
    // Only proceed if consent hasn't been given yet
    if (getSessionRecordingConsent() !== null) return;

    // Check if user is in EU or California
    const requiresConsent = await isEUOrCalifornia();
    if (requiresConsent) {
      showBanner = true;
    } else {
      // Non-EU/CA users: auto-grant consent
      setSessionRecordingConsent("granted");
    }
  });

  function accept() {
    setSessionRecordingConsent("granted");
    showBanner = false;
  }

  function decline() {
    setSessionRecordingConsent("denied");
    showBanner = false;
  }
</script>

{#if showBanner}
  <div class="consent-banner">
    <div class="consent-content">
      <p class="consent-title">Usage Analytics</p>
      <p class="consent-message">
        We collect anonymous usage analytics to improve our app. Sensitive data
        (passwords, emails, phone numbers) is automatically masked and not
        collected.
      </p>
      <p class="consent-disclaimer">
        By accepting, you consent to usage data collection in accordance with
        our
        <a
          href="https://www.rilldata.com/privacy-policy"
          target="_blank"
          rel="noopener noreferrer">Privacy Policy</a
        >. You can change this preference at any time in settings.
      </p>
      <div class="consent-actions">
        <Button type="secondary" onClick={decline} small>Decline</Button>
        <Button type="primary" onClick={accept} small>Accept</Button>
      </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .consent-banner {
    @apply fixed bottom-4 right-4 z-50;
    @apply bg-surface-background border rounded-md shadow-lg;
    @apply max-w-sm p-4;
  }

  .consent-content {
    @apply flex flex-col gap-3;
  }

  .consent-title {
    @apply text-sm font-semibold text-fg-primary;
  }

  .consent-message {
    @apply text-xs text-fg-secondary leading-relaxed;
  }

  .consent-disclaimer {
    @apply text-[11px] text-fg-muted leading-relaxed;
  }

  .consent-disclaimer a {
    @apply text-accent-primary-action underline;
  }

  .consent-actions {
    @apply flex gap-2 justify-end;
  }
</style>
