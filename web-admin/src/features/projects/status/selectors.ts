import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data: { prodDeployment?: V1Deployment }) => {
          // There may not be a prodDeployment if the project is hibernating
          return data?.prodDeployment;
        },
      },
    },
  );
}

// export function useResources(instanceId: string, isAdmin = false) {
//   let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

//   return createRuntimeServiceListResources(
//     instanceId,
//     {
//       // Only skip security checks for admin users
//       skipSecurityChecks: isAdmin,
//     },
//     {
//       query: {
//         select: (data) => ({
//           ...data,
//           // Filter out project parser and refresh triggers
//           resources: data?.resources?.filter(
//             (resource) =>
//               resource.meta.name.kind !== ResourceKind.ProjectParser &&
//               resource.meta.name.kind !== ResourceKind.RefreshTrigger,
//           ),
//         }),
//         refetchInterval: (query) => {
//           const newInterval = pollUntilResourcesReconciled(
//             currentRefetchInterval,
//             query.state.data,
//             query,
//           );
//           if (newInterval === false) {
//             currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
//             return false;
//           }
//           currentRefetchInterval = newInterval;
//           return newInterval;
//         },
//       },
//     },
//   );
// }
