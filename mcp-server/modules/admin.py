import os

import httpx
from fastmcp import FastMCP

RILL_ADMIN_BASE_URL = os.getenv("RILL_ADMIN_BASE_URL") or "https://admin.rilldata.com"
RILL_ORGANIZATION_NAME = os.getenv("RILL_ORGANIZATION_NAME")
RILL_SERVICE_TOKEN = os.getenv("RILL_SERVICE_TOKEN")

headers = {}
headers["Authorization"] = f"Bearer {RILL_SERVICE_TOKEN}"

admin_client = httpx.AsyncClient(
    base_url=RILL_ADMIN_BASE_URL,
    headers=headers,
)

admin_mcp = FastMCP(name="RillAdminServer")


@admin_mcp.tool()
async def list_projects(
    organization_name: str = RILL_ORGANIZATION_NAME,
):
    response = await admin_client.get(f"/v1/organizations/{organization_name}/projects")

    # Extract the project names from the response
    # This prevents us from returning too much data in the response
    names = [project["name"] for project in response.json().get("projects", [])]

    return names


@admin_mcp.tool()
async def get_project_runtime(
    project_name: str,
    organization_name: str = RILL_ORGANIZATION_NAME,
):
    response = await admin_client.get(
        f"/v1/organizations/{organization_name}/projects/{project_name}"
    )

    response_json = response.json()
    prod_deployment = response_json.get("prodDeployment", {})

    return {
        "host": prod_deployment.get("runtimeHost"),
        "instance_id": prod_deployment.get("runtimeInstanceId"),
        "jwt": response_json.get("jwt"),
    }


if __name__ == "__main__":
    admin_mcp.run()
