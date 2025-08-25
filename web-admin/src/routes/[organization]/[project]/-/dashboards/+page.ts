export const load = ({ url: { searchParams } }) => {
  const deploying = searchParams.get("deploying");
  const deployingName = searchParams.get("deployingName");

  return {
    deploying,
    deployingName,
  };
};
