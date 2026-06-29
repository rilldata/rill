import * as m from "@rilldata/web-common/paraglide/messages.js";

export function getOrgRolesDescriptionMap() {
  return {
    admin: m.org_role_admin_description(),
    editor: m.org_role_editor_description(),
    viewer: m.org_role_viewer_description(),
    guest: m.org_role_guest_description(),
  };
}

// Source: https://github.com/rilldata/rill/blob/main/admin/database/validate.go#L57
export const SLUG_REGEX = /^[_a-zA-Z0-9][-_a-zA-Z0-9]*$/;
