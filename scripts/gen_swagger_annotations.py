#!/usr/bin/env python3
"""Generate swag annotations for all API Gateway Echo handlers.

Parses each handler file's route registrations
(routerX.VERB(path, apiHandler.Handle("...", handler.Method))) and inserts a
swag comment block above every handler func. Idempotent: any existing comment
blocks above a handler func are removed first, so re-running converges to one
block per route.

Usage: python3 scripts/gen_swagger_annotations.py
"""
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
HANDLER_DIR = ROOT / "service/apigateway/handler"
RESP_DIR = ROOT / "shared/domain/response"
REQ_DIR = ROOT / "shared/domain/requests"

# Cache of type names actually defined in shared/domain/response/*.go so the
# generator never emits an annotation referencing a nonexistent type (swag
# init hard-fails on those).
_RESP_TYPES = None
_REQ_TYPES = None


def available_response_types():
    global _RESP_TYPES
    if _RESP_TYPES is None:
        names = set()
        for f in RESP_DIR.glob("*.go"):
            names.update(re.findall(r"^type (\w+)", f.read_text(), re.M))
        _RESP_TYPES = names
    return _RESP_TYPES


def available_request_types():
    """Type names in shared/domain/requests, for @Param body validation."""
    global _REQ_TYPES
    if _REQ_TYPES is None:
        names = set()
        for f in REQ_DIR.glob("*.go"):
            names.update(re.findall(r"^type (\w+)", f.read_text(), re.M))
        _REQ_TYPES = names
    return _REQ_TYPES


def safe_type(candidate, fallback):
    """Return candidate if defined in shared/domain/response, else fallback."""
    return candidate if candidate in available_response_types() else fallback


def safe_body_type(candidate):
    """Return candidate only if defined in shared/domain/requests (else None, so
    the @Param body line is skipped — swag hard-fails on unknown types)."""
    return candidate if candidate in available_request_types() else None


GROUP_RE = re.compile(r"(\w+)\s*:?=\s*router\.Group\(\"([^\"]+)\"\)")
ROUTE_RE = re.compile(
    r"(\w+)\.(GET|POST|PUT|PATCH|DELETE)\(\"([^\"]*)\",\s*"
    r"apiHandler\.Handle\(\"[^\"]*\",\s*\w+\.(\w+)\)\)"
)
FUNC_RE = re.compile(r"^(func \(h \*(\w+)\) (\w+)\(c echo\.Context\) error) \{", re.MULTILINE)
REQ_RE = re.compile(r"requests\.(\w+)")

INT_PARAMS = {"id", "user_id", "merchant_id", "category_id", "product_id",
              "cashier_id", "order_id", "order_item_id", "role_id",
              "transaction_id", "year", "month", "page", "pageSize"}

PUBLIC_AUTH = {"verify-code", "forgot-password", "reset-password", "hello",
               "register", "login", "refresh-token"}

AUTH_RESP = {
    "HandleHello": None,
    "VerifyCode": "ApiResponseVerifyCode",
    "ForgotPassword": "ApiResponseForgotPassword",
    "ResetPassword": "ApiResponseResetPassword",
    "Register": "ApiResponseRegister",
    "Login": "ApiResponseLogin",
    "RefreshToken": "ApiResponseRefreshToken",
    "GetMe": "ApiResponseGetMe",
}

STATS = {
    "Category": {
        "FindMonthTotalPrice": "ApiResponseCategoryMonthlyTotalPrice",
        "FindYearTotalPrice": "ApiResponseCategoryYearlyTotalPrice",
        "FindMonthPrice": "ApiResponseCategoryMonthPrice",
        "FindYearPrice": "ApiResponseCategoryYearPrice",
    },
    "Cashier": {
        "FindMonthTotalSales": "ApiResponseCashierMonthlyTotalSales",
        "FindYearTotalSales": "ApiResponseCashierYearlyTotalSales",
        "FindMonthSales": "ApiResponseCashierMonthSales",
        "FindYearSales": "ApiResponseCashierYearSales",
    },
    "Order": {
        "FindMonthlyTotalRevenue": "ApiResponseOrderMonthlyTotalRevenue",
        "FindYearlyTotalRevenue": "ApiResponseOrderYearlyTotalRevenue",
        "FindMonthlyRevenue": "ApiResponseOrderMonthly",
        "FindYearlyRevenue": "ApiResponseOrderYearly",
    },
    "Transaction": {
        "FindMonthlySuccess": "ApiResponsesTransactionMonthSuccess",
        "FindYearlySuccess": "ApiResponsesTransactionYearSuccess",
        "FindMonthlyFailed": "ApiResponsesTransactionMonthFailed",
        "FindYearlyFailed": "ApiResponsesTransactionYearFailed",
        "FindMonthlyMethodSuccess": "ApiResponsesTransactionMonthMethod",
        "FindYearlyMethodSuccess": "ApiResponsesTransactionYearMethod",
        "FindMonthlyMethodFailed": "ApiResponsesTransactionMonthMethod",
        "FindYearlyMethodFailed": "ApiResponsesTransactionYearMethod",
    },
}

TAGS = {
    "authHandleApi": "Auth",
    "userHandleApi": "User",
    "roleHandleApi": "Role",
    "categoryHandleApi": "Category",
    "cashierHandleApi": "Cashier",
    "merchantHandleApi": "Merchant",
    "merchantDocumentHandleApi": "Merchant Document",
    "orderHandleApi": "Order",
    "orderItemHandleApi": "Order Item",
    "productHandleApi": "Product",
    "transactionHandleApi": "Transaction",
}

DOMAIN = {
    "authHandleApi": "Auth", "userHandleApi": "User", "roleHandleApi": "Role",
    "categoryHandleApi": "Category", "cashierHandleApi": "Cashier",
    "merchantHandleApi": "Merchant", "merchantDocumentHandleApi": "MerchantDocument",
    "orderHandleApi": "Order", "orderItemHandleApi": "OrderItem",
    "productHandleApi": "Product", "transactionHandleApi": "Transaction",
}


def response_type(domain, method):
    d = DOMAIN.get(domain, "")
    if domain == "authHandleApi":
        return AUTH_RESP.get(method, "ApiResponse" + d)
    if method == "FindAll" or method.startswith("FindAll"):
        return "ApiResponsePagination" + d
    if method in ("FindByActive", "FindByTrashed"):
        return "ApiResponsePagination" + d + "DeleteAt"
    if method in ("FindById", "FindByUserId"):
        return "ApiResponse" + d
    if method.startswith("FindByMerchant"):
        return "ApiResponses" + d if d == "Product" else "ApiResponsePagination" + d
    if method.startswith("FindByCategory"):
        return "ApiResponses" + d
    for prefix, t in STATS.get(d, {}).items():
        if method.startswith(prefix):
            return t
    if method.startswith("RestoreAll") or method == "RestoreAll":
        return "ApiResponse" + d + "All"
    if method.startswith("DeleteAll") or method == "DeleteAllPermanent":
        return "ApiResponse" + d + "All"
    if "Permanent" in method:
        return "ApiResponse" + d + "Delete"
    if method.startswith("Trash") or method.startswith("Trashed") or method.startswith("Restore"):
        # Some domains (role, merchant_document) lack a single-item DeleteAt
        # envelope — fall back to the plain single-item response.
        return safe_type("ApiResponse" + d + "DeleteAt", "ApiResponse" + d)
    return "ApiResponse" + d


def summary(tag, method, path):
    m = method.lower()
    if "hello" in m:
        return "Health check"
    if "verify" in m:
        return "Verify reset code"
    if "forgot" in m:
        return "Request password reset"
    if "reset" in m:
        return "Reset password"
    if "refresh" in m:
        return "Refresh access token"
    if "register" in m:
        return "Register a new user"
    if "login" in m:
        return "Login"
    if "getme" in m:
        return "Get current user profile"
    if method.startswith("FindAll"):
        return f"List all {tag.lower()} (paginated)"
    if method == "FindByActive":
        return f"List active {tag.lower()}"
    if method == "FindByTrashed":
        return f"List trashed {tag.lower()}"
    if method == "FindById":
        return f"Get {tag.lower()} by ID"
    if method == "FindByUserId":
        return f"Get {tag.lower()} by user ID"
    if method.startswith("FindByMerchant"):
        return f"List {tag.lower()} by merchant"
    if method.startswith("FindByCategory"):
        return "List products by category"
    if method.startswith("FindMonth") or method.startswith("FindYear"):
        return f"Get {tag.lower()} statistics"
    if method.startswith("RestoreAll"):
        return f"Restore all trashed {tag.lower()}"
    if method.startswith("DeleteAll"):
        return f"Delete all trashed {tag.lower()} permanently"
    if "Permanent" in method:
        return f"Delete {tag.lower()} permanently"
    if method.startswith("Trash") or method.startswith("Trashed"):
        return f"Trash {tag.lower()}"
    if method.startswith("Restore"):
        return f"Restore {tag.lower()}"
    if method.startswith("Update") and "Status" in method:
        return f"Update {tag.lower()} status"
    if method.startswith("Update"):
        return f"Update {tag.lower()}"
    if method.startswith("Create"):
        return f"Create {tag.lower()}"
    return method


def is_stats(method):
    return method.startswith("FindMonth") or method.startswith("FindYear")


def is_list(method):
    return (method.startswith("FindAll") or method in ("FindByActive", "FindByTrashed")
            or method.startswith("FindByMerchant") or method.startswith("FindByCategory"))


def param_type(name):
    return "int" if name in INT_PARAMS else "string"


def gen_annotation(tag, domain, verb, path, method, body_type, secured):
    lines = [f"// {summary(tag, method, path)}",
             f"// @Summary {summary(tag, method, path)}",
             f"// @Tags {tag}",
             "// @Accept json",
             "// @Produce json"]
    for pm in re.findall(r":(\w+)", path):
        lines.append(f'// @Param {pm} path {param_type(pm)} true "{pm.replace("_", " ")}"')
    if is_list(method):
        lines.append('// @Param page query int false "Page number"')
        lines.append('// @Param pageSize query int false "Page size"')
        lines.append('// @Param search query string false "Search keyword"')
    if is_stats(method):
        lines.append('// @Param year query int false "Year (e.g. 2026)"')
        lines.append('// @Param month query int false "Month (1-12)"')
    body_type = safe_body_type(body_type) if body_type else None
    if body_type:
        lines.append(f'// @Param request body requests.{body_type} true "Request body"')
    rt = response_type(domain, method)
    if rt is None:
        lines.append('// @Success 200 {string} string "OK"')
    else:
        lines.append(f"// @Success 200 {{object}} response.{rt}")
    lines += ['// @Failure 400 {object} response.ErrorResponse',
              '// @Failure 404 {object} response.ErrorResponse',
              '// @Failure 500 {object} response.ErrorResponse']
    full_path = path if path.startswith("/api") else "/api/" + path.lstrip("/")
    lines.append(f"// @Router {full_path} [{verb.lower()}]")
    if secured:
        lines.append("// @Security BearerAuth")
    return "\n".join(lines)


def is_secured(tag, path):
    if tag == "Auth":
        # PUBLIC_AUTH holds bare endpoint names; compare the last path segment
        # so /api/auth/hello -> "hello" is matched correctly.
        return path.rstrip("/").split("/")[-1] not in PUBLIC_AUTH
    return True


def strip_existing(lines):
    """Drop previously-generated swagger comment lines above every handler
    func (idempotency), preserving human-written doc comments.

    Only ``// @`` swagger lines and blank lines are stripped; ordinary
    comments like ``// Verify reset code`` are kept so regeneration never
    destroys hand-written documentation. Comment lines are appended to ``out``
    as we walk forward, so when a func is reached the already-appended
    trailing swagger/blank lines must be popped before the func line is kept.
    """
    out = []
    for line in lines:
        if FUNC_RE.match(line):
            while out and (out[-1].lstrip().startswith("// @") or out[-1].strip() == ""):
                out.pop()
        out.append(line)
    return out


def process(file_path):
    text = file_path.read_text()
    lines = strip_existing(text.split("\n"))

    groups = {}
    for m in GROUP_RE.finditer(text):
        groups[m.group(1)] = m.group(2)

    routes = {}
    for m in ROUTE_RE.finditer(text):
        g, verb, p, method = m.groups()
        routes.setdefault(method, (verb, groups.get(g, "") + p))

    if not routes:
        return text

    tag = "Api"
    htype = ""
    fm = FUNC_RE.search(text)
    if fm:
        htype = fm.group(2)
        tag = TAGS.get(htype, htype)
    # NB: response_type()/is_secured() key off the RECEIVER type (e.g.
    # "authHandleApi"), not the display tag ("Auth") — pass htype through.
    domain = htype

    body_req = {}
    cur_method = None
    for line in lines:
        m = FUNC_RE.match(line)
        if m:
            cur_method = m.group(3)
            continue
        if cur_method:
            for rm in REQ_RE.finditer(line):
                body_req.setdefault(cur_method, rm.group(1))
            if line.startswith("func "):
                cur_method = None

    out = []
    used = set()
    for line in lines:
        m = FUNC_RE.match(line)
        if m and m.group(3) in routes:
            method = m.group(3)
            if method in used:
                out.append(line)
                continue
            used.add(method)
            verb, path = routes[method]
            out.append(gen_annotation(tag, domain, verb, path, method,
                                      body_req.get(method), is_secured(tag, path)))
        out.append(line)
    return "\n".join(out)


def main():
    total = 0
    for f in sorted(HANDLER_DIR.glob("*.go")):
        if f.name == "handler.go":
            continue
        new = process(f)
        if new != f.read_text():
            f.write_text(new)
            n = len(re.findall(r"^// @Router ", new, re.M))
            total += n
            print(f"{f.name}: {n} annotations")
    print(f"TOTAL annotations: {total}")


if __name__ == "__main__":
    main()
