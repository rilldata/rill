<script lang="ts">
import { slide } from "svelte/transition";
import ColumnEntry from "./ColumnEntry.svelte";
import {DataTypeIcon} from "$lib/components/data-types";
import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import TopKSummary from "$lib/components/viz/TopKSummary.svelte";
import FormattedDataType from "$lib/components/data-types/FormattedDataType.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte"
import SlidingWords from "$lib/components/tooltip/SlidingWords.svelte";
import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
import { config } from "./utils";

import { percentage } from "./utils"
import { formatInteger, formatCompactInteger, standardTimestampFormat } from "$lib/util/formatters"
import { CATEGORICALS, NUMERICS, TIMESTAMPS, DATA_TYPE_COLORS, BOOLEANS } from "$lib/duckdb-data-types";

import Histogram from "$lib/components/viz/histogram/SmallHistogram.svelte";
import TimestampHistogram from "$lib/components/viz/histogram/TimestampHistogram.svelte";
import NumericHistogram from "$lib/components/viz/histogram/NumericHistogram.svelte";
import notificationStore from "$lib/components/notifications/";
import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
import TimestampDetail from "../data-graphic/compositions/timestamp-detail/TimestampDetail.svelte";
import TimestampSpark from "../data-graphic/compositions/timestamp-detail/TimestampSpark.svelte";

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

let active = false;

export function close() {
    active = false;
}
$: exampleWidth = containerWidth > config.mediumCutoff ? config.exampleWidth.medium : config.exampleWidth.small;
$: summaryWidthSize = config.summaryVizWidth[containerWidth < compactBreakpoint ? 'small' : 'medium'];
$: cardinalityFormatter = containerWidth > config.compactBreakpoint ? formatInteger : formatCompactInteger;

let titleTooltip;

function convert(d) {
    return d.map(di => {
        let pi = {...di}
        pi.ts = new Date(pi.ts);
        return pi;
    })
}
</script>

    <!-- pl-10 -->
    <ColumnEntry
    left={indentLevel === 1 ? 10 : 4}
    {hideRight}
    {active}
    emphasize={active}
    on:shift-click={async () => {
        await navigator.clipboard.writeText(name);
        notificationStore.send({ message: `copied column name "${name}" to clipboard`});
    }}
    on:select={async (event) => {
        // we should only allow activation when there are rows present.
        if (totalRows) {
            active = !active;
        }
    }}
>
    <svelte:fragment slot="icon">
        <DataTypeIcon type={type} />
    </svelte:fragment>

    <svelte:fragment slot="left">
        <Tooltip location="right" alignment="center" distance={40} bind:active={titleTooltip}>
            <!-- Wrap in a traditional div then force the ellipsis overflow in the child element.
                this will make the tooltip bound to the parent element while the child element can flow more freely
                and create the ellipisis due to the overflow.
            -->
            <div style:width="100%">
                <div class="column-profile-name text-ellipsis overflow-hidden whitespace-nowrap">
                    {name}
                </div>
            </div>
        <TooltipContent slot="tooltip-content">

            <TooltipTitle>
                <svelte:fragment slot="name">
                    {name}
                </svelte:fragment>
                <svelte:fragment slot="description">
                        {type}
                </svelte:fragment>
            </TooltipTitle>


            {#if totalRows}
                <TooltipShortcutContainer>

                    <SlidingWords {active} hovered={titleTooltip}>
                        {#if CATEGORICALS.has(type)}
                            the top 10 values
                        {:else if TIMESTAMPS.has(type)}
                            the count(*) over time
                        {:else if NUMERICS.has(type)}
                            the distribution of values
                        {/if}
                    </SlidingWords>
                    <Shortcut>
                        Click
                    </Shortcut>

                    <div>
                        <StackingWord>
                            copy
                        </StackingWord>
                        column name to clipboard
                    </div>
                    <Shortcut>
                        <span style='font-family: var(--system);";
                        '>⇧</span> + Click
                    </Shortcut>
                </TooltipShortcutContainer>
            {:else}
                <!-- no data is available, so let's give a useful message-->
                no rows selected
            {/if}
        </TooltipContent>
</Tooltip>
    </svelte:fragment>
    
    <svelte:fragment slot="right">

        <div class="flex gap-2 items-center"  class:hidden={view !== 'summaries'}>
            <div class="flex items-center"  style:width="{summaryWidthSize}px">
                <!-- check to see if the summary has cardinality. Otherwise do not show these values.-->
                {#if totalRows}
                    {#if (CATEGORICALS.has(type) || BOOLEANS.has(type)) && summary?.cardinality}
                        <Tooltip location="right" alignment="center" distance={8} >
                            <BarAndLabel 
                            color={DATA_TYPE_COLORS['VARCHAR'].bgClass}
                            value={summary?.cardinality / totalRows}>
                                |{cardinalityFormatter(summary?.cardinality)}|
                            </BarAndLabel>
                            <TooltipContent slot="tooltip-content" >
                                {formatInteger(summary?.cardinality)} unique values
                            </TooltipContent>
                        </Tooltip>
                    
                    {:else if NUMERICS.has(type) && summary?.histogram?.length}
                    <Tooltip location="right" alignment="center" distance={8}>
                        <Histogram data={summary.histogram} width={summaryWidthSize} height={18} 
                            fillColor={DATA_TYPE_COLORS['DOUBLE'].vizFillClass}
                            baselineStrokeColor={DATA_TYPE_COLORS['DOUBLE'].vizStrokeClass}    
                        />
                        <TooltipContent slot="tooltip-content" >
                            the distribution of the values of this column
                        </TooltipContent>
                    </Tooltip>
                    {:else if TIMESTAMPS.has(type) && 
                        /** a legacy histogram type or a new rollup spark */
                        (summary?.histogram?.length || summary?.rollup?.spark?.length)}
                    <Tooltip location="right" alignment="center" distance={8}>
                        {#if summary?.rollup?.spark}
                            
                            <TimestampSpark 
                                data={convert(summary.rollup.spark)}
                                xAccessor=ts
                                yAccessor=count
                                width={summaryWidthSize}
                                height={18}
                                top={0}
                                bottom={0}
                                left={0}
                                right={0}
                                area
                            />
                        {:else}
                            <Histogram data={summary.histogram} width={summaryWidthSize} height={18} 
                                fillColor={DATA_TYPE_COLORS['TIMESTAMP'].vizFillClass}
                                baselineStrokeColor={DATA_TYPE_COLORS['TIMESTAMP'].vizStrokeClass}    
                            />
                        {/if}
                            <TooltipContent slot="tooltip-content" >
                                the time series
                            </TooltipContent>
                        </Tooltip>
                    {/if}
                {/if}
            </div>

            <div style:width="{config.nullPercentageWidth}px" class:hidden={hideNullPercentage}>

                {#if
                    totalRows !== 0 && 
                    totalRows !== undefined && 
                    nullCount !== undefined}
                <Tooltip location="right" alignment="center" distance={8}>
                    <BarAndLabel
                        showBackground={nullCount !== 0}
                        color={DATA_TYPE_COLORS[type].bgClass}
                        value={nullCount / totalRows || 0}>
                                <span class:text-gray-300={nullCount === 0}>∅ {percentage(nullCount / totalRows)}</span>
                    </BarAndLabel>
                    <TooltipContent slot="tooltip-content" >
                        <svelte:fragment slot="title">
                            what percentage of values are null?
                        </svelte:fragment>
                        {#if nullCount > 0}
                            {percentage(nullCount / totalRows)} of the values are null
                        {:else}
                            no null values in this column
                        {/if}
                    </TooltipContent>
                </Tooltip>
                {/if}

            </div>

        </div>
        <Tooltip location="right" alignment="center" distance={8}>

        <div 
        class:hidden={view !== 'example'}
        class="
            pl-8 text-ellipsis overflow-hidden whitespace-nowrap text-right" style:max-width="{exampleWidth}px"
        >
                <FormattedDataType {type} isNull={example === null || example === ''} value={example} />
        </div>
        <TooltipContent slot="tooltip-content" >
                <FormattedDataType value={example} {type} isNull={example === null || example === ''} dark />
        </TooltipContent>
        </Tooltip>
    </svelte:fragment>

    <svelte:fragment slot="context-button">
        <slot name="context-button">
        </slot>
    </svelte:fragment>

    <svelte:fragment slot="details">
        {#if active}
        <div transition:slide|local={{duration: 200}} class="pt-3 pb-3  w-full">
            {#if (CATEGORICALS.has(type) || BOOLEANS.has(type)) && summary?.topK}
                <div class="pl-{indentLevel ===  1 ? 16 : 10} pr-4 w-full">
                    <!-- pl-16 pl-8 -->
                    <TopKSummary {containerWidth} color={DATA_TYPE_COLORS['VARCHAR'].bgClass} {totalRows} topK={summary.topK} />
                </div>

            {:else if NUMERICS.has(type) && summary?.statistics && summary?.histogram?.length}
            <div class="pl-{indentLevel === 1 ? 12 : 4}">
                <!-- pl-12 pl-5 -->
                <!-- FIXME: we have to remove a bit of pad from the right side to make this work -->
                <NumericHistogram
                    width={containerWidth - (indentLevel === 1 ? (20 + 24 + 44 ): 32)}
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
            {:else if TIMESTAMPS.has(type) && (summary?.histogram?.length || summary.rollup)}
                <div class="pl-{indentLevel === 1 ? 16 : 10}">
                    <!-- pl-14 pl-10 -->
                    {#if summary.rollup}
                        <TimestampDetail 
                            data={summary.rollup.results.map(di => {
                                let pi = {...di};
                                pi.ts = new Date(pi.ts);
                                return pi;
                            })}
                            spark={summary.rollup.spark.map(di => {
                                let pi = {...di};
                                pi.ts = new Date(pi.ts);
                                return pi;
                            })}
                            xAccessor=ts
                            yAccessor=count
                            mouseover={true}
                            height={160}
                            width={containerWidth - (indentLevel === 1 ? (20 + 24 + 54 ): 32 + 20)}
                            rollupGrain={summary.rollup.granularity}
                            estimatedSmallestTimeGrain={summary?.estimatedSmallestTimeGrain}
                            interval={summary.interval}
                        />
                    {:else}
                    <TimestampHistogram
                        {type}
                        width={containerWidth - (indentLevel === 1 ? (20 + 24 + 54 ): 32 + 20)}
                        data={summary.histogram}
                        interval={summary.interval}
                        estimatedSmallestTimeGrain={summary?.estimatedSmallestTimeGrain}
                    />
                    {/if}
                </div>
            {/if}
        </div>
        {/if}
    </svelte:fragment>

</ColumnEntry>
