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
    base_url=f"{RILL_ADMIN_BASE_URL}/v1/orgs/{RILL_ORGANIZATION_NAME}/projects/{RILL_PROJECT_NAME}",
    headers=headers,
)
