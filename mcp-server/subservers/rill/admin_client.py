import os
import sys

import httpx

RILL_ADMIN_BASE_URL = os.getenv("RILL_ADMIN_BASE_URL") or "https://admin.rilldata.com"
RILL_ORGANIZATION_NAME = os.getenv("RILL_ORGANIZATION_NAME")
RILL_PROJECT_NAME = os.getenv("RILL_PROJECT_NAME")
RILL_SERVICE_TOKEN = os.getenv("RILL_SERVICE_TOKEN")

if not RILL_ORGANIZATION_NAME or not RILL_PROJECT_NAME or not RILL_SERVICE_TOKEN:
    print(
        "RILL_ORGANIZATION_NAME, RILL_PROJECT_NAME, and RILL_SERVICE_TOKEN must be set",
        file=sys.stderr,
    )
    raise SystemExit(1)

headers = {}
headers["Authorization"] = f"Bearer {RILL_SERVICE_TOKEN}"

admin_client = httpx.AsyncClient(
    base_url=RILL_ADMIN_BASE_URL,
    headers=headers,
)

runtime_info = None


async def get_project():
    response = await admin_client.get(
        f"/v1/organizations/{RILL_ORGANIZATION_NAME}/projects/{RILL_PROJECT_NAME}"
    )

    return response.json()


def fix_dev_runtime_host(host: str) -> str:
    if host == "http://localhost:9091":
        return "http://localhost:8081"
    return host


async def get_runtime_info(force_refresh=False):
    global runtime_info

    if runtime_info is None or force_refresh:
        project = await get_project()
        prod_deployment = project.get("prodDeployment", {})
        host = fix_dev_runtime_host(prod_deployment.get("runtimeHost"))
        runtime_info = {
            "host": host,
            "instance_id": prod_deployment.get("runtimeInstanceId"),
            "jwt": project.get("jwt"),
        }

    if runtime_info is None:
        raise ValueError("Failed to get runtime info")

    return runtime_info
