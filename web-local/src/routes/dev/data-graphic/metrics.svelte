<script lang="ts">
  import { format } from "d3-format";
  import { cubicOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { derived } from "svelte/store";

  import {
    Body,
    GraphicContext,
    SimpleDataGraphic,
  } from "@rilldata/web-common/components/data-graphic/elements";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import {
    Axis,
    Grid,
    PointLabel,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import type { PointLabelVariant } from "@rilldata/web-common/components/data-graphic/guides/types";
  import {
    Area,
    Line,
  } from "@rilldata/web-common/components/data-graphic/marks";

  function makeData(intervalSize = 1000000) {
    let v1 = 36;
    let v2 = 50;
    const windowSize1 = 1 + ~~(Math.random() * 5);
    const windowSize2 = 1 + ~~(Math.random() * 5);
    const data = Array.from({ length: 50 }).map((_, i) => {
      v1 += 10 * (Math.random() - 0.5);
      v2 += 5 * (Math.random() - 0.5);
      if (v1 < 0) v1 = 0;
      if (v2 < 0) v2 = 0;
      return {
        period: new Date(+new Date("2010-01-01 00:01:04") + i * intervalSize),
        metric1: v1,
        metric2: v2,
        metric3: v1 * v2,
        metric4: v1 * 10 - v2,
      };
    });
    return data
      .map(({ period, metric4 }, i) => {
        const window1 = data.slice(Math.max(0, i - windowSize1), i);
        const window2 = data.slice(Math.max(0, i - windowSize2), i);
        const v = window1.reduce((acc, v) => acc + v.metric1, 0);
        const v2 = window2.reduce((acc, v) => acc + v.metric2, 0);
        return {
          period,
          metric1: v / window1.length,
          metric2: v2 / window2.length,
          metric3: (v * v2) / window1.length,
          metric4: metric4 / 10000,
        };
      })
      .slice(1);
  }
  let data1 = tweened(makeData(), { easing });

  let bigNum1 = derived(
    data1,
    ($data) => {
      return {
        metric1: $data.reduce((a, b) => a + b.metric1, 0),
        metric2: $data.reduce((a, b) => a + b.metric2, 0),
        metric3: $data.reduce((a, b) => a + b.metric3, 0),
        metric4: $data.reduce((a, b) => a + b.metric4, 0),
      };
    },
    { metric1: 0, metric2: 0, metric3: 0, metric4: 0 }
  );

  let metrics = [
    {
      name: "Daily Active Users",
      accessor: "metric1",
      formatBigNumber: format(",.3r"),
    },
    {
      name: "New Signups",
      accessor: "metric2",
      formatBigNumber: format(",.3r"),
      formatAxis: format("~s"),
    },
    {
      name: "Revenue",
      accessor: "metric3",
      formatBigNumber: format("$,.3r"),
      formatAxis: format("~s"),
    },
    {
      name: "Something Else",
      accessor: "metric4",
      formatBigNumber: format(",.2%"),
      formatAxis: format(".0%"),
    },
  ];

  let mouseoverValue = undefined;

  let mouseoverStyle: PointLabelVariant = "fixed";
  function style(style: PointLabelVariant) {
    return () => {
      mouseoverStyle = style;
    };
  }
</script>

<button
  on:click={() => {
    data1.set(makeData());
  }}>randomize</button
>

<div>
  <button
    class:bg-gray-100={mouseoverStyle === "fixed"}
    on:click={style("fixed")}>top right</button
  >
  <button
    class:bg-gray-100={mouseoverStyle === "moving"}
    on:click={style("moving")}>with mouse</button
  >
</div>

<div
  style="
    display: grid;
    grid-template-columns: 140px max-content;
  "
  style:width="max-content"
>
  <GraphicContext
    width={500}
    height={125}
    left={24}
    right={45}
    top={8}
    bottom={4}
    yMin={0}
    bodyBuffer={0}
    xType="date"
    yType="number"
  >
    <div />
    <SimpleDataGraphic top={24} height={24} let:config>
      <Axis side="top" />
    </SimpleDataGraphic>
    <WithBisector
      data={$data1}
      callback={(datum) => datum.period}
      value={mouseoverValue?.period}
      let:point
    >
      {#each metrics as { name, accessor, formatBigNumber, formatAxis }, i (i)}
        <div>
          <h2>
            {name}
          </h2>
          <div
            style:font-size="1.5rem"
            style:font-weight="light"
            class="text-gray-600"
          >
            {formatBigNumber
              ? formatBigNumber($bigNum1[accessor])
              : $bigNum1[accessor]}
          </div>
        </div>
        <SimpleDataGraphic
          shareYScale={false}
          yType="number"
          xType="date"
          yMin={0}
          bind:mouseoverValue
        >
          <Body border borderColor="rgba(0,0,0,.1)">
            <Line data={$data1} xAccessor="period" yAccessor={accessor} />
            <Area data={$data1} xAccessor="period" yAccessor={accessor} />
          </Body>
          <Grid showY={false} />
          <Axis side="right" formatter={formatAxis} />
          <WithBisector
            data={$data1}
            callback={(datum) => datum.period}
            value={mouseoverValue?.x}
            let:point
          >
            {#if point}
              <PointLabel
                tweenProps={{ duration: 50 }}
                variant={mouseoverStyle}
                x={point.period}
                y={point[accessor]}
                format={formatBigNumber}
              />
            {/if}
          </WithBisector>
        </SimpleDataGraphic>
      {/each}
    </WithBisector>
  </GraphicContext>
</div>
