export const load = ({ url: { searchParams } }) => {
  const deploying = searchParams.get("deploying");
  const deployingName = searchParams.get("deploying_name");

  return {
    deploying,
    deployingName,
  };
};
