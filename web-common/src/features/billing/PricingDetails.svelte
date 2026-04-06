<script lang="ts">
  export let extraText = "";
  // "legacy" = existing Team plan copy; "managed" / "live_connect" = new PRD v10 copy
  export let mode: "legacy" | "managed" | "live_connect" = "legacy";
</script>

<div>
  {extraText}
  {#if mode === "legacy"}
    Pricing is based on amount of data ingested (and compressed) into Rill.
  {:else if mode === "managed"}
    Pure usage-based billing with no base fee.
  {:else}
    Cluster Slots are auto-calculated from your OLAP cluster. Add Rill Slots for
    extra performance.
  {/if}
  <ul class="mt-5 ml-5 list-disc">
    {#if mode === "legacy"}
      <li>Starts at $250/month with 10 GB included, $25/GB thereafter</li>
      <li>Unlimited projects, limited to 50 GB each</li>
    {:else if mode === "managed"}
      <li>$0.15/slot/hr (~$110/slot/month always-on)</li>
      <li>$1/GB/month storage above 1GB included</li>
      <li>Minimum 2 slots once over 1GB</li>
    {:else}
      <li>
        Cluster Slots: $0.06/slot/hr (auto-calculated from your OLAP cluster)
      </li>
      <li>Rill Slots: $0.15/slot/hr (user-controlled, starts at 0)</li>
    {/if}
  </ul>
</div>
