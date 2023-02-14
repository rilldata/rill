<script>
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { NicelyFormattedTypes } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import TimeSeriesBody from "@rilldata/web-common/features/dashboards/workspace/metrics-container/TimeSeriesBody.svelte";
  import TimeSeriesChartContainer from "@rilldata/web-common/features/dashboards/workspace/metrics-container/TimeSeriesChartContainer.svelte";

  const startValue = new Date("2020-08-04T18:30:00.000Z");
  const endValue = new Date("2020-08-05T00:29:59.000Z");

  let mouseoverValue = undefined;

  const key = `${startValue}` + `${endValue}`;

  const badData = [
    {
      measure_0: 0,
      ts: new Date("2020-08-04T18:30:00.000Z"),
    },
  ];

  const goodData = [
    {
      measure_0: 135,
      ts: "2020-08-04T18:30:00.000Z",
    },
    {
      measure_0: 125,
      ts: "2020-08-04T19:30:00.000Z",
    },
    {
      measure_0: 149,
      ts: "2020-08-04T20:30:00.000Z",
    },
    {
      measure_0: 132,
      ts: "2020-08-04T21:30:00.000Z",
    },
    {
      measure_0: 184,
      ts: "2020-08-04T22:30:00.000Z",
    },
    {
      measure_0: 74,
      ts: "2020-08-04T23:30:00.000Z",
    },
    {
      measure_0: 60,
      ts: "2020-08-05T00:29:59.000Z",
    },
  ].map((di) => {
    return { ...di, ts: new Date(di.ts) };
  });

  let data = goodData;
  let option = "good";

  function onSelectChange() {
    if (option === "good") data = goodData;
    else data = badData;
  }
</script>

<select bind:value={option} on:change={onSelectChange}>
  <option value="good"> good data </option>
  <option value="bad"> bad data </option>
</select>

<WithBisector
  {data}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer start={startValue} end={endValue}>
    <div style:padding-left="10px" style:padding-top="30px">
      {#if point?.ts}
        <div class="absolute text-gray-600">
          {new Intl.DateTimeFormat("en-US", {
            dateStyle: "medium",
            timeStyle: "medium",
          }).format(point?.ts)}
        </div>
        &nbsp;
      {:else}
        &nbsp;
      {/if}
    </div>
    <div />
    <SimpleDataGraphic
      height={32}
      top={34}
      bottom={0}
      xMin={startValue}
      xMax={endValue}
    >
      <Axis superlabel side="top" />
    </SimpleDataGraphic>
    <div />
    <TimeSeriesBody
      bind:mouseoverValue
      formatPreset={NicelyFormattedTypes.HUMANIZE}
      {data}
      {key}
      mouseover={point}
      accessor="measure_0"
      start={startValue}
      end={endValue}
    />
  </TimeSeriesChartContainer>
</WithBisector>
