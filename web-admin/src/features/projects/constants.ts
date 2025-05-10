export const PROJECT_ROLES_OPTIONS = [
  {
    value: "admin",
    label: "Admin",
    description: "Full control of project settings and members",
  },
  {
    value: "editor",
    label: "Editor",
    description: "Can create and edit dashboards; manage non-admin access",
  },
  {
    value: "viewer",
    label: "Viewer",
    description: "Read-only access to all project resources",
  },
  {
    value: "guest",
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
