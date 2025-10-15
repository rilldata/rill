export async function load({ parent, params, url }) {
  await parent();

  const organization = params.organization;
  const project = params.project;
  const exploreName = url.searchParams.get("explore") ?? "";

  return {
    organization,
    project,
    exploreName,
  };
}
