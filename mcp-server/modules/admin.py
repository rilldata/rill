import os

import httpx
from fastmcp import FastMCP

RILL_ADMIN_BASE_URL = os.getenv("RILL_ADMIN_BASE_URL")
RILL_ADMIN_SERVICE_TOKEN = os.getenv("RILL_ADMIN_SERVICE_TOKEN")
RILL_ADMIN_ORGANIZATION_NAME = os.getenv("RILL_ADMIN_ORGANIZATION_NAME")

headers = {}
headers["Authorization"] = f"Bearer {RILL_ADMIN_SERVICE_TOKEN}"

admin_client = httpx.AsyncClient(
    base_url=RILL_ADMIN_BASE_URL,
    headers=headers,
)

admin_mcp = FastMCP(name="RillAdminServer")


@admin_mcp.tool()
async def get_organization(
    organization_name: str = RILL_ADMIN_ORGANIZATION_NAME,
):
    response = await admin_client.get(f"/v1/organizations/{organization_name}")
    return response.json()


@admin_mcp.tool()
async def list_projects(
    organization_name: str = RILL_ADMIN_ORGANIZATION_NAME,
):
    response = await admin_client.get(f"/v1/organizations/{organization_name}/projects")
    return response.json()


@admin_mcp.tool()
async def get_project(
    project_name: str,
    organization_name: str = RILL_ADMIN_ORGANIZATION_NAME,
):
    response = await admin_client.get(
        f"/v1/organizations/{organization_name}/projects/{project_name}"
    )
    return response.json()


admin_mcp._mcp_server.instructions = (
    "This server provides access to RillData Admin APIs."
    "Use tools to create/update/delete organizations and projects. "
    "Use resources/templates to list/get organizations and projects."
)


if __name__ == "__main__":
    admin_mcp.run()
