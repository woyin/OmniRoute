import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const script = fileURLToPath(new URL("../../scripts/parity/inventory-main-routes.mjs", import.meta.url));

function run(...args) {
  return execFileSync(process.execPath, [script, ...args], { encoding: "utf8" });
}

test("main route inventory is stable, canonical, unique, and contains core routes", () => {
  const first = run("--allow-unknown");
  const second = run("--allow-unknown");
  assert.equal(first, second);

  const inventory = JSON.parse(first);
  assert.equal(inventory.schemaVersion, 1);
  assert.ok(inventory.routes.length > 0);

  const keys = inventory.routes.map(({ method, path }) => `${method} ${path}`);
  assert.equal(new Set(keys).size, keys.length);
  assert.deepEqual(
    inventory.routes,
    [...inventory.routes].sort(
      (a, b) => a.path.localeCompare(b.path) || a.method.localeCompare(b.method),
    ),
  );
  assert.ok(keys.includes("POST /api/auth/login"));
  assert.ok(keys.includes("GET /api/health/ping"));
  assert.ok(keys.includes("POST /api/v1/chat/completions"));
  assert.ok(keys.includes("GET /api/v1/models/{...}"));

  for (const route of inventory.routes) {
    assert.match(route.path, /^\/api(?:\/|$)/);
    assert.doesNotMatch(route.path, /\[/);
    assert.match(route.source, /^main:src\/app\/api\/.+\/route\.ts$/);
    assert.ok(["unknown"].includes(route.auth));
    assert.ok(["unknown"].includes(route.stream));
  }
});

test("strict mode rejects unknown classifications using stderr only", () => {
  const result = spawnSync(process.execPath, [script], { encoding: "utf8" });
  assert.notEqual(result.status, 0);
  assert.equal(result.stdout, "");
  assert.match(result.stderr, /unknown auth\/stream classification/);
});


test("--ref requires a value", () => {
  const result = spawnSync(process.execPath, [script, "--ref"], { encoding: "utf8" });
  assert.notEqual(result.status, 0);
  assert.equal(result.stdout, "");
  assert.match(result.stderr, /missing --ref value/);
});
