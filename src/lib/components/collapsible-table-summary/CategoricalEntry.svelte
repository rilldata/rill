<script lang="ts">
import { slide } from "svelte/transition";
import ColumnEntry from "./ColumnEntry.svelte";
import {DataTypeIcon} from "$lib/components/data-types";
import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import TopKSummary from "$lib/components/viz/TopKSummary.svelte";
import { config } from "./utils";

import { percentage } from "./utils"
import { formatInteger } from "$lib/util/formatters"
import { CATEGORICALS, NUMERICS, TIMESTAMPS } from "$lib/duckdb-data-types";

import Histogram from "$lib/components/viz/SmallHistogram.svelte";
import SummaryAndHistogram from "$lib/components/viz/SummaryAndHistogram.svelte";
import TimestampHistogram from "$lib/components/viz/TimestampHistogram.svelte";
import NumericHistogram from "$lib/components/viz/NumericHistogram.svelte"

export let name;
export let type;
export let summary;
export let totalRows;
export let nullCount;
export let example;
export let view = 'summaries'; // summaries, example
export let containerWidth:number;

export let hideRight = false;

let active = false;

export function close() {
    active = false;
}

</script>

<ColumnEntry
    {hideRight}
    {active}
    emphasize={active}
    on:select={() => {
        active = !active;
    }}
>
    <svelte:fragment slot="icon">
        <DataTypeIcon type={type} />
    </svelte:fragment>

    <svelte:fragment slot="left">
        {name}
    </svelte:fragment>
    
    <svelte:fragment slot="right">
        <div class="flex gap-2">

            <div  style:width={config.summaryVizWidth}>

                {#if CATEGORICALS.has(type)}
                    <BarAndLabel 
                    color={ 'hsl(240, 50%, 90%'}
                    value={summary.cardinality / totalRows}>
                        |{formatInteger(summary.cardinality)}|
                    </BarAndLabel>
                {:else if NUMERICS.has(type) && summary?.histogram}
                    <Histogram data={summary.histogram} width={98} height={19} color={"hsl(1,50%, 80%)"} />
                {:else if TIMESTAMPS.has(type) && summary?.histogram}
                    <Histogram data={summary.histogram} width={98} height={19} color={"#14b8a6"} />
                {/if}

            </div>

            <div style:width={config.nullPercentageWidth}>
                {#if totalRows !== undefined && nullCount !== undefined}
                    <BarAndLabel
                        title="{name}: {percentage(nullCount / totalRows)} of the values are null"
                        bgColor={nullCount === 0 ? 'bg-white' : 'bg-gray-50'}
                        color={'hsl(240, 50%, 90%'}
                        value={nullCount / totalRows || 0}>
                                <span class:text-gray-300={nullCount === 0}>∅ {percentage(nullCount / totalRows)}</span>
                    </BarAndLabel>
                {/if}
            </div>

        </div>
    </svelte:fragment>

    <svelte:fragment slot="context-button">
        <slot name="context-button"></slot>
    </svelte:fragment>

    <svelte:fragment slot="details">
        {#if active}
        <div transition:slide|local={{duration: 200}} class="pt-3 pb-3">
            {#if CATEGORICALS.has(type)}
                <div class="pl-16">
                    <TopKSummary {totalRows} topK={summary.topK} />
                </div>

            {:else if NUMERICS.has(type) && summary?.statistics && summary?.histogram}
            <div class="pl-12">
                <NumericHistogram
                    width={containerWidth - 32}
                    height={65} 
                    data={summary.histogram}
                    min={summary.statistics.min}
                    qlow={summary.statistics.q25}
                    median={summary.statistics.q50}
                    qhigh={summary.statistics.q75}
                    mean={summary.statistics.mean}
                    max={summary.statistics.max}
                />
            </div>
            {:else if TIMESTAMPS.has(type)}
                <div class="pl-14">
                    <TimestampHistogram
                        width={containerWidth - 32 - 20}
                        data={summary.histogram}
                        interval={summary.interval}
                    />
                </div>
            {/if}
        </div>
        {/if}
    </svelte:fragment>

</ColumnEntry>