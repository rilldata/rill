export async function load({ parent, params }) {
  await parent();

  const organization = params.organization;
  const project = params.project;
  const exploreName = params.name;

  return {
    organization,
    project,
    exploreName,
  };
}
