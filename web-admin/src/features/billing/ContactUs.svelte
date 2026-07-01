<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let variant: "normal" | "enterprise" = "normal";

  $: text = variant === "normal" ? m.billing_contact_us_link() : m.billing_contact_support();

  function onClick() {
    window.Pylon("show");
  }
</script>

<button onclick={onClick} class="text-primary-600 text-sm ml-0.5">
  {text}
</button>
