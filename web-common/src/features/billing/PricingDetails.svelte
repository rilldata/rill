<script lang="ts">
  import { PricingDetails } from "@rilldata/web-common/features/billing/pricing-details.ts";

  export let extraText = "";
</script>

<div>
  {extraText} Pricing is based on compute and amount of data ingested (and compressed)
  into Rill.
  <a
    href="https://www.rilldata.com/pricing"
    target="_blank"
    rel="noreferrer noopener"
  >
    See pricing details ->
  </a>

  <ul class="mt-5 ml-5 list-disc">
    <li>{PricingDetails}</li>
    <li>Unlimited projects, limited to 50 GB each</li>
  </ul>
</div>
