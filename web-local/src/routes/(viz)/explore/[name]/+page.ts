export const load = ({ params, depends }) => {
  const exploreName = params.name;

  depends(exploreName, "explore");

  return { exploreName };
};
