<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ApacheDruid from "@rilldata/web-common/components/icons/connectors/ApacheDruid.svelte";
  import ClickHouse from "@rilldata/web-common/components/icons/connectors/ClickHouse.svelte";
  import DuckDb from "@rilldata/web-common/components/icons/connectors/DuckDB.svelte";
  import { createEventDispatcher } from "svelte";
  import { type OlapDriver } from "../../../connectors/olap/olap-config";
  import "./wizard.css";

  const dispatch = createEventDispatcher();

  export let managementType: "rill-managed" | "self-managed";
  export let selectedOLAP: OlapDriver;
  export let onSelectOLAP: (olap: OlapDriver) => void;

  $: olapOptions =
    managementType === "rill-managed"
      ? [
          {
            name: "duckdb",
            icon: DuckDb,
            iconPosition: {
              width: 85,
              height: 24,
              top: 12,
            },
            copy: "Ideal for projects up to 10GB",
          },
          {
            name: "clickhouse",
            icon: ClickHouse,
            iconPosition: {
              width: 108,
              height: 18,
              top: 14,
            },
            copy: "Great for projects up to 100GB",
          },
        ]
      : [
          {
            name: "clickhouse",
            icon: ClickHouse,
            iconPosition: {
              width: 108,
              height: 18,
              top: 14,
            },
            copy: "Great for projects up to 100GB",
          },
          {
            name: "druid",
            icon: ApacheDruid,
            iconPosition: {
              width: 85,
              height: 22,
              top: 12,
            },
            copy: "Connect to an existing cluster",
          },
        ];
</script>

<section class="flex flex-col gap-y-4 items-center">
  <div class="olap-cards">
    {#each olapOptions as option (option.name)}
      <button
        class="option"
        class:selected={selectedOLAP === option.name}
        on:click={() => onSelectOLAP(option.name)}
      >
        <div
          class="absolute"
          style="width: {option.iconPosition.width}px; height: {option
            .iconPosition.height}px; top: {option.iconPosition.top}px;"
        >
          <svelte:component this={option.icon} />
        </div>
        <small class="description">{option.copy}</small>
      </button>
    {/each}
  </div>

  {#if managementType === "self-managed"}
    <Button
      wide
      type="primary"
      disabled={!selectedOLAP}
      on:click={() => dispatch("continue")}
    >
      Continue
    </Button>
  {/if}
</section>

<style lang="postcss">
  .olap-cards {
    @apply flex justify-center gap-x-4;
  }

  button {
    @apply w-[196px] h-[64px] px-4 py-3;
    @apply flex flex-col items-center justify-center relative;
  }

  .description {
    @apply absolute bottom-3;
    @apply text-slate-500 text-xs;
  }
</style>
