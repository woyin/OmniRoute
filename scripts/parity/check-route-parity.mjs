#!/usr/bin/env node

import fs from "node:fs";
import process from "node:process";

import { compareRouteContracts } from "./route-contract-lib.mjs";

function value(flag) {
  const index = process.argv.indexOf(flag);
  if (index < 0 || !process.argv[index + 1]) throw new Error(`missing ${flag}`);
  return process.argv[index + 1];
}

try {
  const main = JSON.parse(fs.readFileSync(value("--main"), "utf8"));
  const go = JSON.parse(fs.readFileSync(value("--go"), "utf8"));
  const maxMissing = Number(value("--max-missing"));
  const maxExtra = Number(value("--max-extra"));
  const maxAuth = Number(value("--max-auth"));
  const maxStream = Number(value("--max-stream"));
  const ceilings = [maxMissing, maxExtra, maxAuth, maxStream];
  if (ceilings.some((ceiling) => !Number.isInteger(ceiling) || ceiling < 0)) {
    throw new Error("parity ceilings must be non-negative integers");
  }
  const report = compareRouteContracts(main.routes, go.routes);
  const knownAuthMismatches = report.authMismatches.filter((mismatch) => mismatch.main !== "unknown");
  const summary = {
    main: main.routes.length,
    go: go.routes.length,
    missing: report.missingInGo.length,
    extra: report.extraInGo.length,
    auth: knownAuthMismatches.length,
    stream: report.streamMismatches.length,
  };
  process.stdout.write(`${JSON.stringify(summary)}\n`);
  for (const route of report.missingInGo) process.stdout.write(`missing: ${route.method} ${route.path}\n`);
  for (const route of report.extraInGo) process.stdout.write(`extra: ${route.method} ${route.path}\n`);
  for (const mismatch of knownAuthMismatches) process.stdout.write(`auth: ${mismatch.method} ${mismatch.path} main=${mismatch.main} go=${mismatch.go}\n`);
  for (const mismatch of report.streamMismatches) process.stdout.write(`stream: ${mismatch.method} ${mismatch.path} main=${mismatch.main} go=${mismatch.go}\n`);

  // ponytail: count ratchet catches growth now; replace with zero-gap exact gate after backlog closes.
  if (summary.missing > maxMissing || summary.extra > maxExtra || summary.auth > maxAuth || summary.stream > maxStream) {
    process.stderr.write(`route parity regressed: missing ${summary.missing}/${maxMissing}, extra ${summary.extra}/${maxExtra}, auth ${summary.auth}/${maxAuth}, stream ${summary.stream}/${maxStream}\n`);
    process.exitCode = 1;
  }
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 1;
}
