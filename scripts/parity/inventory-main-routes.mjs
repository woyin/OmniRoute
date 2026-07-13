#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import process from "node:process";

import { extractRouteMethods, sortRouteContracts } from "./route-contract-lib.mjs";

const args = process.argv.slice(2);
const allowUnknown = args.includes("--allow-unknown");
const refIndex = args.indexOf("--ref");
let ref = refIndex >= 0 ? args[refIndex + 1] : "main";
const missingRef = refIndex >= 0 && (!ref || ref.startsWith("--"));
const unexpected = args.filter((arg, index) =>
  arg !== "--allow-unknown" && arg !== "--ref" && index !== refIndex + 1
);

function git(...args) {
  return execFileSync("git", args, { encoding: "utf8", maxBuffer: 64 * 1024 * 1024 });
}

if (refIndex < 0) {
  try {
    git("rev-parse", "--verify", "main^{commit}");
  } catch {
    ref = "origin/main";
  }
}

function fail(message) {
  process.stderr.write(`${message}\n`);
  process.exitCode = 1;
}

if (missingRef) {
  fail("missing --ref value");
} else if (unexpected.length > 0) {
  fail(`unknown argument: ${unexpected[0]}`);
} else {
  try {
    const files = git("ls-tree", "-r", "--name-only", ref, "--", "src/app/api")
      .split("\n")
      .filter((file) => file.endsWith("/route.ts"));
    const routes = [];

    for (const source of files) {
      const contents = git("show", `${ref}:${source}`);
      const path = `/${source.slice("src/app/".length, -"/route.ts".length)}`;
      for (const method of extractRouteMethods(contents)) {
        routes.push({
          method,
          path,
          source: `${ref}:${source}`,
          auth: "unknown",
          stream: "unknown",
        });
      }
    }

    const sorted = sortRouteContracts(routes);
    const keys = sorted.map(({ method, path }) => `${method} ${path}`);
    const duplicate = keys.find((key, index) => keys.indexOf(key) !== index);
    if (duplicate) {
      fail(`duplicate route contract: ${duplicate}`);
    } else if (!allowUnknown && sorted.some(({ auth, stream }) => auth === "unknown" || stream === "unknown")) {
      fail("unknown auth/stream classification; rerun with --allow-unknown for initial inventory");
    } else {
      process.stdout.write(`${JSON.stringify({ schemaVersion: 1, routes: sorted }, null, 2)}\n`);
    }
  } catch (error) {
    fail(error instanceof Error ? error.message : String(error));
  }
}
