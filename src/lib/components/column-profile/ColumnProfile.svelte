<script lang="ts">
import { slide } from "svelte/transition";
import ColumnEntry from "./ColumnEntry.svelte";
import {DataTypeIcon} from "$lib/components/data-types";
import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import TopKSummary from "$lib/components/viz/TopKSummary.svelte";
import FormattedDataType from "$lib/components/data-types/FormattedDataType.svelte"
import { config } from "./utils";

import { percentage } from "./utils"
import { formatInteger, formatCompactInteger, standardTimestampFormat } from "$lib/util/formatters"
import { CATEGORICALS, NUMERICS, TIMESTAMPS, DATA_TYPE_COLORS } from "$lib/duckdb-data-types";

import Histogram from "$lib/components/viz/histogram/SmallHistogram.svelte";
import TimestampHistogram from "$lib/components/viz/histogram/TimestampHistogram.svelte";
import NumericHistogram from "$lib/components/viz/histogram/NumericHistogram.svelte";

export let name;
export let type;
export let summary;
export let totalRows;
export let nullCount;
export let example;
export let view = 'summaries'; // summaries, example
export let containerWidth:number;

export let indentLevel = 1;

export let hideRight = false;
// hide the null percentage number
export let hideNullPercentage = false;
export let compactBreakpoint = 350;
export let smallSummaryGraphSize = 'medium';

let active = false;

export function close() {
    active = false;
}
$: exampleWidth = containerWidth > 300 ? config.exampleWidth.medium : config.exampleWidth.small;
$: summaryWidthSize = config.summaryVizWidth[containerWidth < compactBreakpoint ? 'small' : 'medium'];
$: cardinalityFormatter = containerWidth > compactBreakpoint ? formatInteger : formatCompactInteger;
</script>

<ColumnEntry
    left={indentLevel === 1 ? 8 : 3}
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

        <div class="flex gap-2 items-center"  class:hidden={view !== 'summaries'}>
            <div class="flex items-center"  style:width="{summaryWidthSize}px">

                {#if CATEGORICALS.has(type)}
                    <BarAndLabel 
                    color={DATA_TYPE_COLORS['VARCHAR'].bgClass}
                    value={summary?.cardinality / totalRows}>
                        |{cardinalityFormatter(summary?.cardinality)}|
                    </BarAndLabel>
                
                {:else if NUMERICS.has(type) && summary?.histogram}
                    <Histogram data={summary.histogram} width={summaryWidthSize} height={18} 
                        fillColor={DATA_TYPE_COLORS['DOUBLE'].vizFillClass}
                        baselineStrokeColor={DATA_TYPE_COLORS['DOUBLE'].vizStrokeClass}    
                    />
                {:else if TIMESTAMPS.has(type) && summary?.histogram}
                    <Histogram data={summary.histogram} width={summaryWidthSize} height={18} 
                        fillColor={DATA_TYPE_COLORS['TIMESTAMP'].vizFillClass}
                        baselineStrokeColor={DATA_TYPE_COLORS['TIMESTAMP'].vizStrokeClass}    
                        />
                {/if}

            </div>

            <div style:width="{config.nullPercentageWidth}px" class:hidden={hideNullPercentage}>

                {#if totalRows !== undefined && nullCount !== undefined}
                    <BarAndLabel
                        title="{name}: {percentage(nullCount / totalRows)} of the values are null"
                        showBackground={nullCount !== 0}
                        color={DATA_TYPE_COLORS[type].bgClass}
                        value={nullCount / totalRows || 0}>
                                <span class:text-gray-300={nullCount === 0}>∅ {percentage(nullCount / totalRows)}</span>
                    </BarAndLabel>
                {/if}

            </div>

        </div>
        <div 
        class:hidden={view !== 'example'}
        class="
            pl-8 text-ellipsis overflow-hidden whitespace-nowrap text-right" style:max-width="{exampleWidth}px"
        >
            <FormattedDataType {type} isNull={example === null}>
                {TIMESTAMPS.has(type) ? standardTimestampFormat(new Date(example)) : example}
                </FormattedDataType>
        </div>
    </svelte:fragment>

    <svelte:fragment slot="context-button">
        <slot name="context-button"></slot>
    </svelte:fragment>

    <svelte:fragment slot="details">
        {#if active}
        <div transition:slide|local={{duration: 200}} class="pt-3 pb-3">
            {#if CATEGORICALS.has(type)}
                <div class="pl-{indentLevel ===  1 ? 16 : 8} pr-8">
                    <!-- pl-16 pl-8 -->
                    <TopKSummary color={DATA_TYPE_COLORS['VARCHAR'].bgClass} {totalRows} topK={summary.topK} />
                </div>

            {:else if NUMERICS.has(type) && summary?.statistics && summary?.histogram}
            <div class="pl-{indentLevel === 1 ? 12 : 5}">
                <!-- pl-12 pl-5 -->
                <!-- FIXME: we have to remove a bit of pad from the right side to make this work -->
                <NumericHistogram
                    width={containerWidth - (indentLevel === 1 ? (20 + 24 + 32 ): 32)}
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
                <div class="pl-{indentLevel === 1 ? 14 : 10}">
                    <!-- pl-14 pl-10 -->
                    <TimestampHistogram
                        width={containerWidth - (indentLevel === 1 ? (20 + 24 + 32 ): 32 + 20)}
                        data={summary.histogram}
                        interval={summary.interval}
                    />
                </div>
            {/if}
        </div>
        {/if}
    </svelte:fragment>

</ColumnEntry>