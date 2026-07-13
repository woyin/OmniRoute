import assert from "node:assert/strict";
import test from "node:test";

import {
  compareRouteContracts,
  extractRouteMethods,
  normalizeRoutePath,
  sortRouteContracts,
} from "../../scripts/parity/route-contract-lib.mjs";

test("normalizeRoutePath canonicalizes Next and chi parameters", () => {
  assert.equal(normalizeRoutePath("/api/users/[id]"), "/api/users/{id}");
  assert.equal(normalizeRoutePath("/api/files/[...catchAll]"), "/api/files/{catchAll...}");
  assert.equal(normalizeRoutePath("/api/users/{id}"), "/api/users/{id}");
});

test("extractRouteMethods finds function, const, and re-export methods", () => {
  const source = `
export async function GET() {}
export const POST = handler;
export { remove as DELETE } from "./handlers";
export function OPTIONS() {}
export const runtime = "nodejs";
`;

  assert.deepEqual(extractRouteMethods(source), ["DELETE", "GET", "POST"]);
});

test("extractRouteMethods preserves explicit uncommon HTTP methods", () => {
  assert.deepEqual(extractRouteMethods("export function PATCH() {}\nexport const HEAD = handler;"), [
    "HEAD",
    "PATCH",
  ]);
});

test("sortRouteContracts orders by path then method without mutating input", () => {
  const routes = [
    { method: "POST", path: "/z", auth: "required", stream: "json" },
    { method: "GET", path: "/a", auth: "none", stream: "sse" },
    { method: "DELETE", path: "/a", auth: "required", stream: "json" },
  ];

  assert.deepEqual(
    sortRouteContracts(routes).map(({ path, method }) => `${path} ${method}`),
    ["/a DELETE", "/a GET", "/z POST"],
  );
  assert.equal(routes[0].path, "/z");
});

test("compareRouteContracts reports route, auth, and stream differences stably", () => {
  const main = [
    { method: "POST", path: "/v1/chat", auth: "required", stream: "sse" },
    { method: "GET", path: "/health", auth: "none", stream: "json" },
    { method: "OPTIONS", path: "/ignored", auth: "none", stream: "json" },
    { method: "GET", path: "/missing", auth: "required", stream: "json" },
  ];
  const go = [
    { method: "GET", path: "/health", auth: "required", stream: "json" },
    { method: "POST", path: "/v1/chat", auth: "required", stream: "json" },
    { method: "DELETE", path: "/extra", auth: "required", stream: "json" },
  ];

  assert.deepEqual(compareRouteContracts(main, go), {
    missingInGo: [{ method: "GET", path: "/missing", auth: "required", stream: "json" }],
    extraInGo: [{ method: "DELETE", path: "/extra", auth: "required", stream: "json" }],
    authMismatches: [
      {
        method: "GET",
        path: "/health",
        main: "none",
        go: "required",
      },
    ],
    streamMismatches: [
      {
        method: "POST",
        path: "/v1/chat",
        main: "sse",
        go: "json",
      },
    ],
  });
});
