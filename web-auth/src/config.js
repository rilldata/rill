import Google from "@rilldata/web-common/components/icons/Google.svelte";
import Microsoft from "@rilldata/web-common/components//icons/Microsoft.svelte";
import Okta from "@rilldata/web-common/components//icons/Okta.svelte";
import Pingfed from "@rilldata/web-common/components//icons/Pingfed.svelte";

export const LOGIN_OPTIONS = [
  {
    name: "Google",
    icon: Google,
    connection: "google-oauth2",
    label: "Continue with Google",
    style: "primary",
  },
  {
    name: "Microsoft",
    icon: Microsoft,
    connection: "windowslive",
    label: "Continue with Microsoft",
    style: "secondary",
  },
  {
    name: "Okta",
    icon: Okta,
    label: "Continue with Okta",
    style: "secondary",
  },
  {
    name: "Pingfed",
    icon: Pingfed,
    label: "Continue with Ping Fed",
    style: "secondary",
  },
];
