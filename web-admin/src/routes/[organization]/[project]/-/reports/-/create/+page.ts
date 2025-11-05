export async function load({ parent, params, url: { searchParams } }) {
  await parent();

  const organization = params.organization;
  const project = params.project;

  const exploreName = searchParams.get("explore") ?? "";
  const canvasName = searchParams.get("canvas") ?? "";

  const aggregationRequest =
    JSON.parse(searchParams.get("query") || "{}")[
      "metricsViewAggregationRequest"
    ] ?? {};
  const metricsViewName = aggregationRequest.metricsView;

  return {
    organization,
    project,
    metricsViewName,
    exploreName,
    canvasName,
    aggregationRequest,
  };
}
