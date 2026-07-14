import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export function getProjectRolesOptions() {
  return [
    {
      value: ProjectUserRoles.Admin,
      label: m.project_share_role_admin(),
      description: m.project_share_role_admin_description(),
    },
    {
      value: ProjectUserRoles.Editor,
      label: m.project_share_role_editor(),
      description: m.project_share_role_editor_description(),
    },
    {
      value: ProjectUserRoles.Viewer,
      label: m.project_share_role_viewer(),
      description: m.project_share_role_viewer_description(),
    },
  ];
}

export function getProjectRolesDescriptionMap() {
  return {
    admin: m.project_share_role_admin_description(),
    editor: m.project_share_role_editor_description(),
    viewer: m.project_share_role_viewer_description(),
    guest: m.project_role_guest_description(),
  };
}
