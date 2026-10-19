#!/usr/bin/env python3
"""Synthetic local smoke test for company scope and memberships.

Only reserved @zu360.test accounts and fixed synthetic company IDs are accepted.
"""

import getpass
import json
import os
import sys
import urllib.error
import urllib.request

BASE_URL = os.environ.get("ZU360_BASE_URL", "http://127.0.0.1:4001").rstrip("/")
COMPANY_A1 = "22222222-2222-4222-8222-222222222222"
COMPANY_A2 = "66666666-6666-4666-8666-666666666666"
COMPANY_B1 = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
COMPANY_B2 = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
EMAIL_A = os.environ.get("TEST_SST_A_EMAIL", "sst-a@zu360.test")
EMAIL_B = os.environ.get("TEST_SST_B_EMAIL", "sst-b@zu360.test")


def call(method: str, path: str, *, token: str | None = None, payload: dict | None = None, headers: dict | None = None) -> tuple[int, dict]:
    body = json.dumps(payload).encode() if payload is not None else None
    request_headers = {"Accept": "application/json"}
    if body is not None:
        request_headers["Content-Type"] = "application/json"
    if token:
        request_headers["Authorization"] = f"Bearer {token}"
    if headers:
        request_headers.update(headers)
    request = urllib.request.Request(BASE_URL + path, data=body, headers=request_headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as error:
        error.read()  # consume response without printing server details
        return error.code, {}


def login(email: str, password: str) -> str:
    status, response = call("POST", "/v1/auth/login", payload={"email": email, "password": password})
    if status != 200 or not response.get("access_token"):
        raise RuntimeError(f"login failed for synthetic account ({email}); HTTP {status}")
    return response["access_token"]


def switch_company(token: str, company_id: str) -> tuple[int, str, str]:
    status, response = call("POST", "/v1/auth/switch-company", token=token, payload={"company_id": company_id})
    return status, response.get("access_token", ""), response.get("active_company_id", "")


def expect(label: str, actual: int, expected: int) -> bool:
    passed = actual == expected
    print(f"{'PASS' if passed else 'FAIL'} {label}: expected HTTP {expected}, received {actual}")
    return passed


def main() -> int:
    if not EMAIL_A.endswith("@zu360.test") or not EMAIL_B.endswith("@zu360.test"):
        print("ERROR: smoke test accepts only reserved @zu360.test accounts", file=sys.stderr)
        return 2
    print(f"Testing only reserved synthetic accounts against {BASE_URL}")
    password_a = getpass.getpass(f"Password for {EMAIL_A}: ")
    password_b = getpass.getpass(f"Password for {EMAIL_B}: ")
    try:
        token_a = login(EMAIL_A, password_a)
        token_b = login(EMAIL_B, password_b)
    except (RuntimeError, urllib.error.URLError) as error:
        print(str(error), file=sys.stderr)
        return 2
    finally:
        password_a = ""
        password_b = ""

    checks: list[bool] = []
    path_a1 = f"/v1/companies/{COMPANY_A1}"
    path_a2 = f"/v1/companies/{COMPANY_A2}"
    path_b1 = f"/v1/companies/{COMPANY_B1}"
    path_b2 = f"/v1/companies/{COMPANY_B2}"

    # Membership listing must include the professional's own two companies only.
    status, company_list = call("GET", "/v1/companies", token=token_a)
    checks.append(expect("A lists only membership-authorized companies", status, 200))
    listed_ids = {row.get("id") for row in company_list.get("data", [])}
    checks.append(listed_ids == {COMPANY_A1, COMPANY_A2})
    print(f"{'PASS' if checks[-1] else 'FAIL'} A membership set is exactly A1+A2")
    status, company_list_b = call("GET", "/v1/companies", token=token_b)
    checks.append(expect("B lists only membership-authorized companies", status, 200))
    listed_ids_b = {row.get("id") for row in company_list_b.get("data", [])}
    checks.append(listed_ids_b == {COMPANY_B1, COMPANY_B2})
    print(f"{'PASS' if checks[-1] else 'FAIL'} B membership set is exactly B1+B2")

    # Signed company scope is active; headers/query/body cannot change it.
    status, response = call("GET", path_a1 + "?company_id=" + COMPANY_B1, token=token_a,
                            headers={"X-Company-ID": COMPANY_B1, "X-Tenant-ID": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"})
    checks.append(expect("A accesses active A1; forged header/query are ignored", status, 200))
    checks.append(response.get("id") == COMPANY_A1)
    print(f"{'PASS' if checks[-1] else 'FAIL'} response remains scoped to A1")
    checks.append(expect("A→B1 company request is denied", call("GET", path_b1, token=token_a)[0], 403))
    checks.append(expect("A→B1 worker request is denied", call("GET", path_b1 + "/employees", token=token_a)[0], 403))
    checks.append(expect("A can read synthetic workers in active A1", call("GET", path_a1 + "/employees", token=token_a)[0], 200))

    # A2 is authorized but cannot be used until Core issues an A2-scoped token.
    checks.append(expect("A cannot use A2 before switching active company", call("GET", path_a2, token=token_a)[0], 403))
    status, token_a2, active_id = switch_company(token_a, COMPANY_A2)
    checks.append(expect("A switches to membership-authorized A2", status, 200))
    checks.append(active_id == COMPANY_A2 and bool(token_a2))
    print(f"{'PASS' if checks[-1] else 'FAIL'} Core issued token scoped to A2")
    checks.append(expect("A2 token accesses A2", call("GET", path_a2, token=token_a2)[0], 200))
    checks.append(expect("A2 token reads synthetic workers in A2", call("GET", path_a2 + "/employees", token=token_a2)[0], 200))
    checks.append(expect("A cannot switch to B1 without membership", switch_company(token_a2, COMPANY_B1)[0], 403))

    checks.append(expect("B accesses own B1", call("GET", path_b1, token=token_b)[0], 200))
    checks.append(expect("B reads synthetic workers in own B1", call("GET", path_b1 + "/employees", token=token_b)[0], 200))
    checks.append(expect("B→A2 company request is denied", call("GET", path_a2, token=token_b)[0], 403))
    checks.append(expect("B→A2 worker request is denied", call("GET", path_a2 + "/employees", token=token_b)[0], 403))
    checks.append(expect("B cannot use B2 before switching active company", call("GET", path_b2, token=token_b)[0], 403))
    status, token_b2, active_id_b = switch_company(token_b, COMPANY_B2)
    checks.append(expect("B switches to membership-authorized B2", status, 200))
    checks.append(active_id_b == COMPANY_B2 and bool(token_b2))
    print(f"{'PASS' if checks[-1] else 'FAIL'} Core issued token scoped to B2")
    checks.append(expect("B2 token accesses B2", call("GET", path_b2, token=token_b2)[0], 200))
    checks.append(expect("B2 token reads synthetic workers in B2", call("GET", path_b2 + "/employees", token=token_b2)[0], 200))
    checks.append(expect("B2 token cannot access A1", call("GET", path_a1, token=token_b2)[0], 403))

    token_a = token_a2 = token_b = token_b2 = ""
    if all(checks):
        print("PASS: synthetic membership and company-isolation HTTP smoke test")
        return 0
    print("FAIL: do not use real data; inspect Core logs and the failed gate", file=sys.stderr)
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
