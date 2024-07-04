import {
  createAdminServiceGetProject,
  createAdminServiceListProjectMembers,
} from "@rilldata/web-admin/client";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

export function useProjectMembersEmails(organization: string, project: string) {
  return createAdminServiceListProjectMembers(
    organization,
    project,
    undefined,
    {
      query: {
        select: (data) => {
          return data.members
            ?.filter((member) => !!member?.userEmail)
            .map((member) => member.userEmail);
        },
      },
    },
  );
}

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(
    orgName,
    projectName,
    {},
    {
      query: {
        enabled: !!orgName && !!projectName,
        select: (resp) => resp.project?.id,
      },
    },
  );
}
