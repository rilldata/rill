<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";

  import {
    formatMsInterval,
    formatMsToDuckDbIntervalString,
  } from "../strategies/intervals";

  const MS = 1;
  const SEC = 1000 * MS;
  const MIN = 60 * SEC;
  const HOUR = 60 * MIN;
  const DAY = 24 * HOUR;
  const MONTH = 30 * DAY; //eslint-disable-line
  const YEAR = 365 * DAY; //eslint-disable-line

  const time_formulas = [
    "-2234 * YEAR",
    "-39 * SEC",
    "0",
    "0.0011797",
    "0.01231",
    "1.797",
    "123.7989",
    "793.987",
    "100.9797",
    "1 * SEC",
    "1.4709879 * SEC",
    "9.49797 * SEC",
    "10 * SEC",
    "59 * SEC",
    "1 * MIN",
    "99.9 * SEC",
    "100 * SEC",
    "59.23451 * MIN",
    "89.411 * MIN",
    "89.94353 * MIN",
    "99 * MIN",
    "99.9 * MIN",
    "100 * MIN",
    "71.936 * HOUR",
    "72 * HOUR",
    "99 * HOUR",
    "89.9 * DAY",
    "90 * DAY",
    "99 * DAY",
    "7.87978 * MONTH",
    "17.923 * MONTH",
    "18 * MONTH",
    "18.0234234 * MONTH",
    "36 * MONTH",
    "3247 * DAY",
    "43.34523 * YEAR",
    "99 * YEAR",
    "99 * YEAR + 6 * SEC",
    "99 * YEAR + 6.0004 * SEC",
    "99 * YEAR + 6.99999 * SEC",
    "99.9 * YEAR",
    "100.234 * YEAR",
    "123797.239797 * YEAR",
  ];

  const ms_values = time_formulas.map((f) => ({ string: f, num: eval(f) }));
</script>

<Meta
  title="Data formatting/Intervals"
  argTypes={{
    clickAction: { action: "subbutton-click" },
  }}
/>

<Story name="Intervals (compact)">
  <table>
    <tr style="border-bottom: solid 1px #ddd;">
      <td> input formula</td>
      <td> milliseconds</td>
      <td> formatted interval (compact)</td>
    </tr>
    {#each ms_values as t}
      <tr>
        <td> <pre>{t.string}</pre></td>
        <td> <pre>{t.num}</pre></td>
        <td>
          {formatMsInterval(t.num)}
        </td>
      </tr>
    {/each}
  </table>
</Story>

<Story name="Intervals (extended)">
  <table>
    <tr style="border-bottom: solid 1px #ddd;">
      <!-- <td> input formula</td> -->
      <!-- <td> milliseconds</td> -->
      <td> formatted time (shortest)</td>
      <td> formatted time (units)</td>
      <td> formatted time (colon)</td>
      <td> (same formatting)</td>
    </tr>
    {#each ms_values as t}
      <tr>
        <!-- <td> <pre>{t.string}</pre></td> -->
        <!-- <td> <pre>{t.num}</pre></td> -->
        <td>
          {formatMsToDuckDbIntervalString(t.num)}
        </td>
        <td>{formatMsToDuckDbIntervalString(t.num, "units")} </td>
        <td>{formatMsToDuckDbIntervalString(t.num, "colon")} </td>
        <td
          >{formatMsToDuckDbIntervalString(t.num, "units") ==
          formatMsToDuckDbIntervalString(t.num, "colon")
            ? "==="
            : ""}
        </td>
      </tr>
    {/each}
  </table>
</Story>

<style>
  td {
    padding-left: 1.5em;
  }
</style>
