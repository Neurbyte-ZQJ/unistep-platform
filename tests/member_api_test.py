#!/usr/bin/env python3
"""tests/member_api_test.py

针对团员发展模块的 HTTP 集成测试脚本。
使用方法：

    # 启动后端
    cd backend && go run ./main.go

    # 在另一个终端运行
    python tests/member_api_test.py http://localhost:8080

脚本会：
1. 注册一个临时用户并登录获取 JWT
2. 通过 /api/v1/members 完成 CRUD
3. 依次录入入团申请、积极分子、发展对象、政审备案
4. 调用 /archive 校验时间线长度
5. 校验 /swagger.json 可访问

注意：附件上传依赖 MinIO，未配置 MINIO_ENDPOINT 时会跳过该用例。
"""
from __future__ import annotations

import io
import json
import os
import sys
import time
import urllib.error
import urllib.request
import uuid

BASE = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"
PASS = FAIL = 0


def request(method: str, path: str, body=None, headers=None, raw_body=None, content_type="application/json"):
    url = f"{BASE}{path}"
    hdrs = {"Accept": "application/json"}
    if headers:
        hdrs.update(headers)
    if raw_body is not None:
        data = raw_body
        hdrs["Content-Type"] = content_type
    elif body is not None:
        data = json.dumps(body).encode()
        hdrs["Content-Type"] = "application/json"
    else:
        data = None
    req = urllib.request.Request(url, data=data, headers=hdrs, method=method)
    try:
        resp = urllib.request.urlopen(req)
        text = resp.read().decode()
        return resp.status, json.loads(text) if text else None
    except urllib.error.HTTPError as e:
        text = e.read().decode()
        return e.code, json.loads(text) if text else None


def test(name: str):
    def wrap(fn):
        global PASS, FAIL
        try:
            fn()
            print(f"  PASS: {name}")
            PASS += 1
        except AssertionError as ex:
            print(f"  FAIL: {name} - {ex}")
            FAIL += 1
        return fn
    return wrap


print(f"Testing {BASE}\n")
suffix = uuid.uuid4().hex[:6]
username = f"tester_{suffix}"
student_no = f"S{int(time.time())}"
token = ""
profile_id = 0


@test("注册并登录获取 JWT")
def _():
    global token
    status, _ = request("POST", "/api/v1/auth/register", {
        "username": username, "password": "password123", "email": f"{username}@x.com"
    })
    assert status in (200, 201, 400), f"register status={status}"
    status, body = request("POST", "/api/v1/auth/login", {
        "username": username, "password": "password123"
    })
    assert status == 200, body
    token = body["data"]["token"]
    assert token, "missing token"


def auth():
    return {"Authorization": f"Bearer {token}"}


@test("创建团员档案")
def _():
    global profile_id
    status, body = request("POST", "/api/v1/members", {
        "name": "测试用户",
        "studentNo": student_no,
        "gender": "男",
        "college": "计算机学院",
    }, headers=auth())
    assert status == 201, body
    profile_id = body["data"]["id"]
    assert profile_id > 0


@test("学号重复返回 400")
def _():
    status, body = request("POST", "/api/v1/members", {
        "name": "另一人",
        "studentNo": student_no,
    }, headers=auth())
    assert status == 400, body


@test("列表查询包含新档案")
def _():
    status, body = request("GET", f"/api/v1/members?page=1&size=10&name=测试", headers=auth())
    assert status == 200, body
    assert body["data"]["total"] >= 1


@test("更新档案")
def _():
    status, body = request("PUT", f"/api/v1/members/{profile_id}", {
        "name": "测试用户",
        "studentNo": student_no,
        "phone": "13800000000",
        "stage": "activist",
    }, headers=auth())
    assert status == 200, body


@test("入团申请")
def _():
    status, body = request("POST", f"/api/v1/members/{profile_id}/applications", {
        "applyDate": "2024-09-01", "motivation": "...", "introducer": "辅导员",
    }, headers=auth())
    assert status == 201, body


@test("积极分子培养")
def _():
    status, body = request("POST", f"/api/v1/members/{profile_id}/activists", {
        "startDate": "2024-10-01", "trainer": "李老师", "score": 88,
    }, headers=auth())
    assert status == 201, body


@test("发展对象")
def _():
    status, body = request("POST", f"/api/v1/members/{profile_id}/develop-targets", {
        "confirmedDate": "2025-03-01", "conclusion": "公示通过",
    }, headers=auth())
    assert status == 201, body


@test("政审备案")
def _():
    status, body = request("POST", f"/api/v1/members/{profile_id}/political-reviews", {
        "reviewDate": "2025-04-01", "reviewer": "校团委", "conclusion": "符合条件",
    }, headers=auth())
    assert status == 201, body


@test("生成电子档案")
def _():
    status, body = request("GET", f"/api/v1/members/{profile_id}/archive", headers=auth())
    assert status == 200, body
    summary = body["data"]["summary"]
    assert summary["applicationCount"] == 1
    assert summary["activistCount"] == 1
    assert summary["developRecordCount"] == 1
    assert summary["politicalRecordCount"] == 1
    assert len(body["data"]["timeline"]) == 4


@test("Swagger JSON 可访问")
def _():
    status, body = request("GET", "/swagger.json")
    assert status == 200
    assert body["info"]["title"] == "UniStep Platform API"


if os.environ.get("MINIO_ENDPOINT"):
    @test("附件上传到 MinIO")
    def _():
        import urllib.parse
        boundary = f"----unistep{uuid.uuid4().hex}"
        body = io.BytesIO()
        body.write(f"--{boundary}\r\nContent-Disposition: form-data; name=\"category\"\r\n\r\napplication\r\n".encode())
        body.write(f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"a.txt\"\r\n".encode())
        body.write(b"Content-Type: text/plain\r\n\r\nhello\r\n")
        body.write(f"--{boundary}--\r\n".encode())
        status, resp = request(
            "POST", f"/api/v1/members/{profile_id}/attachments",
            raw_body=body.getvalue(),
            content_type=f"multipart/form-data; boundary={boundary}",
            headers=auth(),
        )
        assert status == 201, resp

print(f"\nResults: {PASS} passed, {FAIL} failed")
sys.exit(0 if FAIL == 0 else 1)
