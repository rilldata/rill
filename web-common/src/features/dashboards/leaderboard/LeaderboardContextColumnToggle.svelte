<script lang="ts">
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import {
    ButtonGroup,
    SubButton,
  } from "@rilldata/web-common/components/button-group";
  import { runtime } from "../../../runtime-client/runtime-store";

  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";

  export let metricViewName: string;
  export let validPercentOfTotal: boolean;

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;
  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let disabledButtons: ("delta" | "pie")[] = [];
  $: {
    disabledButtons = [];
    if (
      !hasTimeSeries ||
      metricsExplorer.selectedComparisonTimeRange === undefined
    )
      disabledButtons.push("delta");
    if (validPercentOfTotal !== true) disabledButtons.push("pie");
  }

  let selectedButton: "delta" | "pie" | null = null;
  // NOTE: time comparison takes precedence over percent of total
  $: selectedButton = metricsExplorer?.showComparison
    ? "delta"
    : metricsExplorer?.showPercentOfTotal
    ? "pie"
    : null;

  const handleContextValueButtonGroupClick = (evt) => {
    const value = evt.detail;
    if (value === "delta" && selectedButton == "delta") {
      metricsExplorerStore.displayComparison(metricViewName, false);
    } else if (value === "delta" && selectedButton != "delta") {
      metricsExplorerStore.displayComparison(metricViewName, true);
    } else if (value === "pie" && selectedButton == "pie") {
      metricsExplorerStore.displayPercentOfTotal(metricViewName, false);
    } else if (value === "pie" && selectedButton != "pie") {
      metricsExplorerStore.displayPercentOfTotal(metricViewName, true);
    }
  };

  $: selectedButtons = selectedButton === null ? [] : [selectedButton];

  const pieTooltips = {
    selected: "Hide percent of total",
    unselected: "Show percent of total",
    disabled:
      "To show percent of total, select a metric that is defined as summable",
  };

  const deltaTooltips = {
    selected: "Hide percent change",
    unselected: "Show percent change",
    disabled: "To show percent change, select a comparison time range",
  };
</script>

<ButtonGroup
  selected={selectedButtons}
  disabled={disabledButtons}
  on:subbutton-click={handleContextValueButtonGroupClick}
>
  <SubButton value={"delta"} tooltips={deltaTooltips}>
    <Delta />%
  </SubButton>
  <SubButton value={"pie"} tooltips={pieTooltips}>
    <PieChart />%
  </SubButton>
</ButtonGroup>
