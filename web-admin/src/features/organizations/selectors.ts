import {
  adminServiceGetOrganization,
  adminServiceListDeployments,
  adminServiceListProjectsForOrganization,
  createAdminServiceListDeployments,
  createAdminServiceListProjectsForOrganization,
  getAdminServiceGetOrganizationQueryKey,
  getAdminServiceListDeploymentsQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
  type V1GetOrganizationResponse,
  type V1Organization,
  type V1Project,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { FetchQueryOptions } from "@tanstack/query-core";
import { derived } from "svelte/store";
import { isActiveDeployment } from "@rilldata/web-admin/features/branches/deployment-utils.ts";

export function areAllProjectsHibernating(organization: string) {
  const projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organization,
      },
    },
    queryClient,
  );
  return derived(projectsQuery, (projectsResp, set) => {
    const isPending = projectsResp.isPending;
    const error = projectsResp.error;
    if (isPending || error) {
      set({
        isPending,
        error,
        data: undefined,
      });
      return;
    }

    const projects = projectsResp.data?.projects ?? [];
    const allHibernating = allProjectsHibernating(projects);
    if (!allHibernating) {
      set({
        isPending: false,
        error: undefined,
        data: false,
      });
      return;
    }

    projects.sort((a, b) => (a.updatedOn > b.updatedOn ? -1 : 1));
    const deploymentQueries = projects
      .slice(0, TopProjectsCount)
      .map((project) =>
        createAdminServiceListDeployments(
          organization,
          project.name,
          {},
          {},
          queryClient,
        ),
      );

    return derived(deploymentQueries, (deploymentResps) => {
      const isPending = deploymentResps.some((dr) => dr.isPending);
      const error = deploymentResps.find((dr) => dr.error)?.error;
      if (isPending || error) {
        return {
          isPending,
          error,
          data: undefined,
        };
      }

      const hasSomeActiveDeployment = deploymentResps.some((deploymentsResp) =>
        deploymentsResp.data?.deployments?.some(
          (d) => d && isActiveDeployment(d),
        ),
      );
      return {
        isPending: false,
        error: error,
        data: !hasSomeActiveDeployment,
      };
    }).subscribe(set);
  });
}

const TopProjectsCount = 5;

export async function fetchAllProjectsHibernating(organization: string) {
  const projectsResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListProjectsForOrganizationQueryKey(organization),
    queryFn: () => adminServiceListProjectsForOrganization(organization),
    staleTime: Infinity,
  });
  const projects = projectsResp.projects ?? [];
  const allHibernating = allProjectsHibernating(projects);
  if (!allHibernating) return false;

  projects.sort((a, b) => (a.updatedOn > b.updatedOn ? -1 : 1));
  const deploymentsQueryPromises = projects
    .slice(0, TopProjectsCount)
    .map(async (project) => {
      const deploymentsResp = await queryClient.fetchQuery({
        queryKey: getAdminServiceListDeploymentsQueryKey(
          organization,
          project.name,
          {},
        ),
        queryFn: () =>
          adminServiceListDeployments(organization, project.name, {}),
        staleTime: Infinity,
      });
      return deploymentsResp.deployments?.some(
        (d) => d && isActiveDeployment(d),
      );
    });
  const hasSomeActiveDeployment = (
    await Promise.all(deploymentsQueryPromises)
  ).some(Boolean);
  return !hasSomeActiveDeployment;
}

export function getFetchOrganizationQueryOptions(
  organization: string | V1Organization | undefined,
) {
  const orgName = normalizeOrganization(organization);
  return <FetchQueryOptions<V1GetOrganizationResponse>>{
    queryKey: getAdminServiceGetOrganizationQueryKey(orgName),
    queryFn: () => adminServiceGetOrganization(orgName),
    staleTime: Infinity,
  };
}

function normalizeOrganization(
  organization: string | V1Organization | undefined,
): string {
  if (typeof organization === "string") {
    return organization;
  }
  if (
    organization &&
    typeof organization === "object" &&
    "name" in organization &&
    typeof organization.name === "string"
  ) {
    return organization.name;
  }
  throw new Error(
    `Invalid organization parameter: expected string or V1Organization object with name property, got ${typeof organization}`,
  );
}

function allProjectsHibernating(projects: V1Project[] | undefined) {
  return projects?.length && projects.every((p) => !p.primaryDeploymentId);
}
