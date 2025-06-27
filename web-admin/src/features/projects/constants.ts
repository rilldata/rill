import { UserRoles } from "@rilldata/web-common/features/users/roles.ts";

export const PROJECT_ROLES_OPTIONS = [
  {
    value: UserRoles.Admin,
    label: "Admin",
    description: "Full control of project settings and members",
  },
  {
    value: UserRoles.Editor,
    label: "Editor",
    description: "Can create and edit dashboards; manage non-admin access",
  },
  {
    value: UserRoles.Viewer,
    label: "Viewer",
    description: "Read-only access to all project resources",
  },
  {
    value: UserRoles.Guest,
    label: "Guest",
    description: "Access to invited projects only",
  },
];

export const PROJECT_ROLES_DESCRIPTION_MAP = {
  admin: "Full control of project settings and members",
  editor: "Can create and edit dashboards; manage non-admin access",
  viewer: "Read-only access to all project resources",
  guest: "Access to invited projects only",
};
