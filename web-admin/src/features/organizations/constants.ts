import { OrgUserRoles } from "@rilldata/web-common/features/users/roles";

export const ORG_ROLES_OPTIONS = [
  {
    value: OrgUserRoles.Admin,
    label: "Admin",
    description: "Full control over organization settings and members",
  },
  {
    value: OrgUserRoles.Editor,
    label: "Editor",
    description: "Can manage projects and most org resources",
  },
  {
    value: OrgUserRoles.Viewer,
    label: "Viewer",
    description: "Read-only access to organization and projects",
  },
];
