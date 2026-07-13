import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const script = fileURLToPath(new URL("../../scripts/parity/check-route-parity.mjs", import.meta.url));

function run(main, go, maxMissing, maxExtra) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "route-parity-"));
  fs.writeFileSync(path.join(dir, "main.json"), JSON.stringify({ routes: main }));
  fs.writeFileSync(path.join(dir, "go.json"), JSON.stringify({ routes: go }));
  return spawnSync(process.execPath, [script, "--main", path.join(dir, "main.json"), "--go", path.join(dir, "go.json"), "--max-missing", String(maxMissing), "--max-extra", String(maxExtra)], { encoding: "utf8" });
}

const route = (method, path) => ({ method, path, auth: "none", stream: "json" });

test("route parity ratchet passes at ceiling and rejects growth", () => {
  const main = [route("GET", "/api/a"), route("POST", "/api/b")];
  const go = [route("GET", "/api/a"), route("GET", "/api/c")];
  const pass = run(main, go, 1, 1);
  assert.equal(pass.status, 0);
  assert.match(pass.stdout, /"missing":1/);
  assert.match(pass.stdout, /missing: POST \/api\/b/);
  const fail = run(main, go, 0, 1);
  assert.notEqual(fail.status, 0);
  assert.match(fail.stderr, /route parity regressed/);
});
