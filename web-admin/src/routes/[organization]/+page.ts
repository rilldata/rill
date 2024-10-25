export const load = async ({ parent }) => {
  const { organizationPermissions } = await parent();
  return {
    organizationPermissions,
  };
};
