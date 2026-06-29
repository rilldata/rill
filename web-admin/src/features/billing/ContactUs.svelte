<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let variant: "normal" | "enterprise" = "normal";

  $: text = variant === "normal" ? m.billing_contact_us_link() : m.billing_contact_support();

  function onClick() {
    window.Pylon("show");
  }
</script>

<button onclick={onClick} class="text-primary-600 text-sm ml-0.5">
  {text}
</button>
