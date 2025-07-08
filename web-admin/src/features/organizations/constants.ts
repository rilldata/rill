export const ORG_ROLES_OPTIONS = [
  {
    value: "admin",
    label: "Admin",
    description: "Full access to org settings, members, and all projects",
  },
  {
    value: "editor",
    label: "Editor",
    description: "Can create/manage projects and non-admin members",
  },
  {
    value: "viewer",
    label: "Viewer",
    description: "Read-only access to all org projects",
  },
  {
    value: "guest",
    label: "Guest",
    description: "Access to invited projects only",
  },
];

export const ORG_ROLES_DESCRIPTION_MAP = {
  admin: "Full access to org settings, members, and all projects",
  editor: "Can create/manage projects and non-admin members",
  viewer: "Read-only access to all org projects",
  guest: "Access to invited projects only",
};

// Source: https://github.com/rilldata/rill/blob/main/admin/database/validate.go#L57
export const SLUG_REGEX = /^[_a-zA-Z0-9][-_a-zA-Z0-9]*$/;
