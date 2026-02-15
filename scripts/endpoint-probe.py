#!/usr/bin/env python3
import json
import os
import ssl
import sys
import urllib.error
import urllib.parse
import urllib.request
from http.cookiejar import CookieJar
from typing import Any, Dict, List, Optional, Tuple


ENDPOINTS: List[Dict[str, Any]] = [
    {"method": "POST", "path": "/api/auth/login"},
    {"method": "GET", "path": "/api/auth/session"},
    {"method": "GET", "path": "/api/settings/general"},
    {"method": "POST", "path": "/api/settings/general"},
    {"method": "GET", "path": "/api/auth/settings"},
    {"method": "PUT", "path": "/api/auth/settings"},
    {"method": "GET", "path": "/api/auth/providers"},
    {"method": "GET", "path": "/api/license"},
    {"method": "POST", "path": "/api/license"},
    {"method": "DELETE", "path": "/api/license"},
    {"method": "GET", "path": "/api/activity"},
    {"method": "GET", "path": "/api/hawser/connect"},
    {"method": "GET", "path": "/api/schedules"},
    {"method": "GET", "path": "/api/schedules/executions"},
    {"method": "POST", "path": "/api/schedules/system/{id}/toggle"},
    {"method": "POST", "path": "/api/schedules/{type}/{id}/toggle"},
    {"method": "POST", "path": "/api/schedules/{type}/{id}/run"},
    {"method": "GET", "path": "/api/users"},
    {"method": "POST", "path": "/api/users"},
    {"method": "GET", "path": "/api/users/{id}"},
    {"method": "PUT", "path": "/api/users/{id}"},
    {"method": "DELETE", "path": "/api/users/{id}"},
    {"method": "GET", "path": "/api/environments"},
    {"method": "POST", "path": "/api/environments"},
    {"method": "GET", "path": "/api/environments/{id}"},
    {"method": "PUT", "path": "/api/environments/{id}"},
    {"method": "DELETE", "path": "/api/environments/{id}"},
    {"method": "GET", "path": "/api/registries"},
    {"method": "POST", "path": "/api/registries"},
    {"method": "GET", "path": "/api/registries/{id}"},
    {"method": "PUT", "path": "/api/registries/{id}"},
    {"method": "DELETE", "path": "/api/registries/{id}"},
    {"method": "GET", "path": "/api/git/credentials"},
    {"method": "POST", "path": "/api/git/credentials"},
    {"method": "GET", "path": "/api/git/credentials/{id}"},
    {"method": "PUT", "path": "/api/git/credentials/{id}"},
    {"method": "DELETE", "path": "/api/git/credentials/{id}"},
    {"method": "GET", "path": "/api/git/repositories"},
    {"method": "POST", "path": "/api/git/repositories"},
    {"method": "GET", "path": "/api/git/repositories/{id}"},
    {"method": "PUT", "path": "/api/git/repositories/{id}"},
    {"method": "DELETE", "path": "/api/git/repositories/{id}"},
    {"method": "GET", "path": "/api/config-sets"},
    {"method": "POST", "path": "/api/config-sets"},
    {"method": "GET", "path": "/api/config-sets/{id}"},
    {"method": "PUT", "path": "/api/config-sets/{id}"},
    {"method": "DELETE", "path": "/api/config-sets/{id}"},
    {"method": "GET", "path": "/api/notifications"},
    {"method": "POST", "path": "/api/notifications"},
    {"method": "GET", "path": "/api/notifications/{id}"},
    {"method": "PUT", "path": "/api/notifications/{id}"},
    {"method": "DELETE", "path": "/api/notifications/{id}"},
    {"method": "GET", "path": "/api/stacks", "with_env": True},
    {"method": "POST", "path": "/api/stacks", "with_env": True},
    {"method": "POST", "path": "/api/stacks/{name}/start", "with_env": True},
    {"method": "POST", "path": "/api/stacks/{name}/stop", "with_env": True},
    {"method": "POST", "path": "/api/stacks/{name}/restart", "with_env": True},
    {"method": "POST", "path": "/api/stacks/{name}/down", "with_env": True},
    {"method": "DELETE", "path": "/api/stacks/{name}", "with_env": True},
    {"method": "GET", "path": "/api/stacks/{name}/env", "with_env": True},
    {"method": "PUT", "path": "/api/stacks/{name}/env", "with_env": True},
    {"method": "GET", "path": "/api/stacks/{name}/env/raw", "with_env": True},
    {"method": "PUT", "path": "/api/stacks/{name}/env/raw", "with_env": True},
    {"method": "POST", "path": "/api/stacks/scan"},
    {"method": "POST", "path": "/api/stacks/adopt"},
    {"method": "GET", "path": "/api/stacks/sources"},
    {"method": "POST", "path": "/api/git/stacks/{id}/webhook"},
    {"method": "POST", "path": "/api/git/stacks/{id}/deploy-stream"},
    {"method": "GET", "path": "/api/git/stacks/{id}/env-files"},
    {"method": "POST", "path": "/api/git/stacks/{id}/env-files"},
    {"method": "GET", "path": "/api/dashboard/stats", "with_env": True},
    {"method": "GET", "path": "/api/networks", "with_env": True},
    {"method": "POST", "path": "/api/networks", "with_env": True},
    {"method": "GET", "path": "/api/networks/{id}/inspect", "with_env": True},
    {"method": "DELETE", "path": "/api/networks/{id}", "with_env": True},
    {"method": "POST", "path": "/api/networks/{id}/connect", "with_env": True},
    {"method": "POST", "path": "/api/networks/{id}/disconnect", "with_env": True},
    {"method": "GET", "path": "/api/volumes", "with_env": True},
    {"method": "POST", "path": "/api/volumes", "with_env": True},
    {"method": "GET", "path": "/api/volumes/{name}/inspect", "with_env": True},
    {"method": "DELETE", "path": "/api/volumes/{name}", "with_env": True},
    {"method": "POST", "path": "/api/volumes/{name}/clone", "with_env": True},
    {"method": "GET", "path": "/api/images", "with_env": True},
    {"method": "POST", "path": "/api/images/pull", "with_env": True},
    {"method": "DELETE", "path": "/api/images/{id}", "with_env": True},
    {"method": "POST", "path": "/api/images/push", "with_env": True},
    {"method": "POST", "path": "/api/images/scan", "with_env": True},
    {"method": "GET", "path": "/api/containers", "with_env": True},
    {"method": "POST", "path": "/api/containers", "with_env": True},
    {"method": "GET", "path": "/api/containers/{id}", "with_env": True},
    {"method": "DELETE", "path": "/api/containers/{id}", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/start", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/stop", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/restart", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/pause", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/unpause", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/rename", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/update", "with_env": True},
    {"method": "GET", "path": "/api/containers/{id}/logs", "with_env": True},
    {"method": "GET", "path": "/api/containers/{id}/top", "with_env": True},
    {"method": "GET", "path": "/api/containers/{id}/shells", "with_env": True},
    {"method": "POST", "path": "/api/containers/{id}/files/create", "with_env": True},
    {"method": "GET", "path": "/api/containers/{id}/files/content", "with_env": True},
    {"method": "PUT", "path": "/api/containers/{id}/files/content", "with_env": True},
    {"method": "DELETE", "path": "/api/containers/{id}/files/delete", "with_env": True},
    {"method": "GET", "path": "/api/containers/stats", "with_env": True},
    {"method": "POST", "path": "/api/containers/check-updates", "with_env": True},
    {"method": "GET", "path": "/api/containers/pending-updates", "with_env": True},
    {"method": "GET", "path": "/api/configs"},
    {"method": "GET", "path": "/api/backups"},
]


def env(name: str, default: Optional[str] = None) -> str:
    value = os.getenv(name, default)
    if value is None:
        raise RuntimeError(f"missing required environment variable: {name}")
    return value


def try_json(text: str) -> Any:
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        return None


def first_id(items: Any, keys: List[str]) -> Optional[str]:
    if not isinstance(items, list) or not items:
        return None
    first = items[0]
    if not isinstance(first, dict):
        return None
    for key in keys:
        if key in first and first[key] is not None:
            return str(first[key])
    return None


class Session:
    def __init__(self, endpoint: str, insecure: bool) -> None:
        self.endpoint = endpoint.rstrip("/")
        cj = CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
        if endpoint.startswith("https://") and insecure:
            ctx = ssl.create_default_context()
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE
            self.opener.add_handler(urllib.request.HTTPSHandler(context=ctx))

    def request(
        self, method: str, path: str, body: Optional[Dict[str, Any]] = None, query: Optional[Dict[str, str]] = None
    ) -> Tuple[int, str]:
        url = self.endpoint + path
        if query:
            url += "?" + urllib.parse.urlencode(query)
        data = None
        if body is not None:
            data = json.dumps(body).encode("utf-8")
        req = urllib.request.Request(url, data=data, method=method)
        req.add_header("Content-Type", "application/json")
        try:
            with self.opener.open(req, timeout=20) as resp:
                payload = resp.read().decode("utf-8", errors="replace")
                return resp.status, payload
        except urllib.error.HTTPError as exc:
            payload = exc.read().decode("utf-8", errors="replace")
            return exc.code, payload


def main() -> int:
    endpoint = env("DOCKHAND_ENDPOINT")
    username = env("DOCKHAND_USERNAME")
    password = env("DOCKHAND_PASSWORD")
    auth_provider = env("DOCKHAND_AUTH_PROVIDER", "local")
    default_env = env("DOCKHAND_DEFAULT_ENV", "1")
    insecure = os.getenv("DOCKHAND_INSECURE", "false").lower() in ("1", "true", "yes")
    allow_mutation = os.getenv("DOCKHAND_PROBE_ALLOW_MUTATION", "false").lower() in ("1", "true", "yes")

    s = Session(endpoint, insecure)
    code, login_body = s.request(
        "POST",
        "/api/auth/login",
        {"username": username, "password": password, "authProvider": auth_provider},
    )
    if code < 200 or code >= 300:
        sys.stderr.write(f"login failed with status {code}: {login_body}\n")
        return 1

    fixtures: Dict[str, Optional[str]] = {
        "env": default_env,
        "name": None,
        "id": None,
        "type": "system",
        "system_schedule_id": None,
        "custom_schedule_id": None,
        "custom_schedule_type": None,
        "user_id": None,
        "environment_id": None,
        "registry_id": None,
        "git_credential_id": None,
        "git_repository_id": None,
        "config_set_id": None,
        "notification_id": None,
        "stack_name": None,
        "git_stack_id": None,
        "network_id": None,
        "volume_name": None,
        "image_id": None,
        "container_id": None,
    }

    # Fixture discovery from safe list endpoints.
    for key, path, keys in [
        ("user_id", "/api/users", ["id", "_id"]),
        ("environment_id", "/api/environments", ["id", "_id"]),
        ("registry_id", "/api/registries", ["id", "_id"]),
        ("git_credential_id", "/api/git/credentials", ["id", "_id"]),
        ("git_repository_id", "/api/git/repositories", ["id", "_id"]),
        ("config_set_id", "/api/config-sets", ["id", "_id"]),
        ("notification_id", "/api/notifications", ["id", "_id"]),
    ]:
        sc, body = s.request("GET", path)
        if sc == 200:
            fixtures[key] = first_id(try_json(body), keys)

    sc, body = s.request("GET", "/api/schedules")
    if sc == 200:
        schedules = try_json(body)
        if isinstance(schedules, dict):
            schedules = schedules.get("schedules", [])
        if isinstance(schedules, list):
            for item in schedules:
                if not isinstance(item, dict):
                    continue
                sid = item.get("id")
                stype = item.get("type")
                if sid is None or stype is None:
                    continue
                if str(stype) == "system" and fixtures["system_schedule_id"] is None:
                    fixtures["system_schedule_id"] = str(sid)
                if str(stype) != "system" and fixtures["custom_schedule_id"] is None:
                    fixtures["custom_schedule_id"] = str(sid)
                    fixtures["custom_schedule_type"] = str(stype)

    sc, body = s.request("GET", "/api/stacks", query={"env": default_env})
    if sc == 200:
        stacks = try_json(body)
        if isinstance(stacks, dict):
            stacks = stacks.get("stacks", [])
        if isinstance(stacks, list):
            if stacks and isinstance(stacks[0], dict):
                if "name" in stacks[0]:
                    fixtures["stack_name"] = str(stacks[0]["name"])
                if "id" in stacks[0]:
                    fixtures["git_stack_id"] = str(stacks[0]["id"])

    for key, path, field in [
        ("network_id", "/api/networks", "id"),
        ("volume_name", "/api/volumes", "name"),
        ("image_id", "/api/images", "id"),
        ("container_id", "/api/containers", "id"),
    ]:
        sc, body = s.request("GET", path, query={"env": default_env})
        if sc == 200:
            items = try_json(body)
            if isinstance(items, list) and items and isinstance(items[0], dict) and field in items[0]:
                fixtures[key] = str(items[0][field])

    rows: List[Dict[str, Any]] = []
    for ep in ENDPOINTS:
        method = ep["method"]
        raw_path = ep["path"]
        path = raw_path
        has_placeholder = ("{id}" in raw_path) or ("{name}" in raw_path) or ("{type}" in raw_path)

        used_fixture = True
        if "{name}" in path:
            name = fixtures.get("stack_name")
            if name and method == "GET":
                path = path.replace("{name}", urllib.parse.quote(name, safe=""))
            else:
                path = path.replace("{name}", "_probe_name_")
                if method == "GET":
                    used_fixture = False
        if "{id}" in path:
            candidate = None
            if "/api/users/" in path:
                candidate = fixtures.get("user_id")
            elif "/api/environments/" in path:
                candidate = fixtures.get("environment_id")
            elif "/api/registries/" in path:
                candidate = fixtures.get("registry_id")
            elif "/api/git/credentials/" in path:
                candidate = fixtures.get("git_credential_id")
            elif "/api/git/repositories/" in path:
                candidate = fixtures.get("git_repository_id")
            elif "/api/config-sets/" in path:
                candidate = fixtures.get("config_set_id")
            elif "/api/notifications/" in path:
                candidate = fixtures.get("notification_id")
            elif "/api/git/stacks/" in path:
                candidate = fixtures.get("git_stack_id")
            elif "/api/networks/" in path:
                candidate = fixtures.get("network_id")
            elif "/api/volumes/" in path:
                candidate = fixtures.get("volume_name")
            elif "/api/images/" in path:
                candidate = fixtures.get("image_id")
            elif "/api/containers/" in path:
                candidate = fixtures.get("container_id")
            elif "/api/schedules/system/" in path:
                candidate = fixtures.get("system_schedule_id")
            elif "/api/schedules/" in path:
                candidate = fixtures.get("custom_schedule_id") or fixtures.get("system_schedule_id")
            if candidate and method == "GET":
                path = path.replace("{id}", urllib.parse.quote(candidate, safe=""))
            else:
                path = path.replace("{id}", "_probe_id_")
                if method == "GET":
                    used_fixture = False
        if "{type}" in path:
            stype = fixtures.get("custom_schedule_type") or "system"
            path = path.replace("{type}", urllib.parse.quote(stype, safe=""))

        if method == "GET" and not used_fixture:
            rows.append(
                {
                    "method": method,
                    "path": raw_path,
                    "http_code": "",
                    "result": "unverified_no_fixture",
                    "note": "No fixture available for path placeholder",
                }
            )
            continue

        query = None
        if ep.get("with_env"):
            query = {"env": default_env}
        payload = None
        if method in ("POST", "PUT"):
            payload = {}
            if raw_path == "/api/auth/login":
                payload = {"username": username, "password": password, "authProvider": auth_provider}

        request_method = method
        note = ""
        # Safe mode: avoid mutating calls unless explicitly enabled.
        if not allow_mutation and method in ("POST", "PUT", "DELETE") and raw_path != "/api/auth/login":
            if has_placeholder:
                # Probe route match without touching real resources by using placeholder ids.
                note = "Safe mode: placeholder probe"
            else:
                # For singleton mutating endpoints, use OPTIONS to avoid side effects.
                request_method = "OPTIONS"
                payload = None
                note = "Safe mode: options probe (mutation skipped)"

        code, body = s.request(request_method, path, payload, query)

        if code == 404 and "{" not in raw_path:
            result = "not_present"
        elif code == 404 and "{" in raw_path:
            result = "unexpected_404"
        else:
            result = "present"

        rows.append(
            {
                "method": method,
                "path": raw_path,
                "http_code": str(code),
                "result": result,
                "note": note,
            }
        )

    out_dir = os.path.join("docs", "reports")
    os.makedirs(out_dir, exist_ok=True)
    out_csv = os.path.join(out_dir, "endpoint-probe.csv")
    out_md = os.path.join(out_dir, "endpoint-probe.md")

    with open(out_csv, "w", encoding="utf-8") as f:
        f.write("method,path,http_code,result,note\n")
        for r in rows:
            f.write(
                f"{r['method']},{r['path']},{r['http_code']},{r['result']},{r['note'].replace(',', ';')}\n"
            )

    total = len(rows)
    present = sum(1 for r in rows if r["result"] == "present")
    missing = sum(1 for r in rows if r["result"] == "not_present")
    unverified = sum(1 for r in rows if r["result"] == "unverified_no_fixture")
    unexpected_404 = sum(1 for r in rows if r["result"] == "unexpected_404")

    with open(out_md, "w", encoding="utf-8") as f:
        f.write("# Endpoint Probe Report\n\n")
        f.write("Generated by `scripts/endpoint-probe.py`.\n\n")
        f.write(f"- Total endpoints: {total}\n")
        f.write(f"- Present (non-404): {present}\n")
        f.write(f"- Not present (404 on non-parameterized route): {missing}\n")
        f.write(f"- Unverified (missing fixture for parameterized route): {unverified}\n")
        f.write(f"- Unexpected 404 (parameterized route with fixture): {unexpected_404}\n\n")
        if missing:
            f.write("## Not Present\n\n")
            for r in rows:
                if r["result"] == "not_present":
                    f.write(f"- `{r['method']} {r['path']}`\n")
            f.write("\n")
        if unverified:
            f.write("## Unverified (No Fixture)\n\n")
            for r in rows:
                if r["result"] == "unverified_no_fixture":
                    f.write(f"- `{r['method']} {r['path']}`\n")
            f.write("\n")
        if unexpected_404:
            f.write("## Unexpected 404\n\n")
            for r in rows:
                if r["result"] == "unexpected_404":
                    f.write(f"- `{r['method']} {r['path']}`\n")
            f.write("\n")

    print(f"wrote {out_csv}")
    print(f"wrote {out_md}")
    print(f"summary: total={total} present={present} not_present={missing} unverified={unverified} unexpected_404={unexpected_404}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
