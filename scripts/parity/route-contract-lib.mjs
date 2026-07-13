const HTTP_METHODS = new Set([
  "CONNECT",
  "DELETE",
  "GET",
  "HEAD",
  "OPTIONS",
  "PATCH",
  "POST",
  "PUT",
  "TRACE",
]);

export function normalizeRoutePath(path) {
  return path
    .replace(/\[\[\.\.\.[^\]]+\]\]/g, "{...?}")
    .replace(/\[\.\.\.[^\]]+\]/g, "{...}")
    .replace(/\[[^\]]+\]/g, "{}")
    .replace(/\{([^}]*)\}/g, (_, parameter) => {
      if (parameter === "...?" || parameter.endsWith("...?")) return "{...?}";
      return parameter === "..." || parameter.endsWith("...") ? "{...}" : "{}";
    });
}

function maskCommentsAndStrings(source) {
  let result = "";
  let state = "code";
  for (let i = 0; i < source.length; i += 1) {
    const char = source[i];
    const next = source[i + 1];
    if (state === "code" && char === "/" && next === "/") {
      state = "line-comment";
      result += "  ";
      i += 1;
    } else if (state === "code" && char === "/" && next === "*") {
      state = "block-comment";
      result += "  ";
      i += 1;
    } else if (state === "code" && (char === '"' || char === "'" || char === "`")) {
      state = char;
      result += " ";
    } else if (state === "line-comment" && char === "\n") {
      state = "code";
      result += char;
    } else if (state === "block-comment" && char === "*" && next === "/") {
      state = "code";
      result += "  ";
      i += 1;
    } else if ((state === '"' || state === "'" || state === "`") && char === "\\") {
      result += "  ";
      i += 1;
    } else if (state === char && (state === '"' || state === "'" || state === "`")) {
      state = "code";
      result += " ";
    } else {
      result += state === "code" ? char : char === "\n" ? "\n" : " ";
    }
  }
  return result;
}

export function extractRouteMethods(source) {
  const methods = new Set();
  const code = maskCommentsAndStrings(source);
  const declarations = /\bexport\s+(?:async\s+)?(?:function|const)\s+([A-Z]+)\b/g;
  const reExports = /\bexport\s*\{([^}]*)\}/g;

  for (const match of code.matchAll(declarations)) {
    if (HTTP_METHODS.has(match[1]) && match[1] !== "OPTIONS") methods.add(match[1]);
  }
  for (const block of code.matchAll(reExports)) {
    for (const match of block[1].matchAll(/\bas\s+([A-Z]+)\b/g)) {
      if (HTTP_METHODS.has(match[1]) && match[1] !== "OPTIONS") methods.add(match[1]);
    }
  }

  return [...methods].sort();
}

export function sortRouteContracts(routes) {
  return routes
    .filter(({ method }) => method !== "OPTIONS")
    .map((route) => ({ ...route, path: normalizeRoutePath(route.path) }))
    .sort((a, b) => a.path.localeCompare(b.path) || a.method.localeCompare(b.method));
}

export function compareRouteContracts(mainRoutes, goRoutes) {
  const main = sortRouteContracts(mainRoutes);
  const go = sortRouteContracts(goRoutes);
  const key = ({ method, path }) => `${method} ${path}`;
  const mainByKey = new Map(main.map((route) => [key(route), route]));
  const goByKey = new Map(go.map((route) => [key(route), route]));

  const missingInGo = main.filter((route) => !goByKey.has(key(route)));
  const extraInGo = go.filter((route) => !mainByKey.has(key(route)));
  const authMismatches = [];
  const streamMismatches = [];

  for (const mainRoute of main) {
    const goRoute = goByKey.get(key(mainRoute));
    if (!goRoute) continue;
    if (mainRoute.auth !== goRoute.auth) {
      authMismatches.push({
        method: mainRoute.method,
        path: mainRoute.path,
        main: mainRoute.auth,
        go: goRoute.auth,
      });
    }
    if (mainRoute.stream !== goRoute.stream) {
      streamMismatches.push({
        method: mainRoute.method,
        path: mainRoute.path,
        main: mainRoute.stream,
        go: goRoute.stream,
      });
    }
  }

  return { missingInGo, extraInGo, authMismatches, streamMismatches };
}

export function classifyMainRouteAuth(source) {
  if (/\brequireManagementAuth\b|\brequireManagementSession\b/.test(source)) return "required";
  if (/\bisAuthenticated\s*\(\s*request\s*\)/.test(source)) return "required";
  if (/\bextractApiKey\b/.test(source) && /\bisRequireApiKeyEnabled\b/.test(source)) return "optional";
  return "unknown";
}

export function classifyMainRouteStream(source, path, method = "GET") {
  if (/\/ws(?:\/|$)|codex-responses-ws|traffic-inspector\/ws/.test(path)) return "websocket";
  if (method === "GET" && (path === "/api/mcp/sse" || path === "/api/mcp/stream")) return "sse";
  if (path === "/api/v1/chat/completions" || path === "/api/v1/responses" || path.startsWith("/api/v1/responses/") || path === "/api/v1/messages") return "json+sse";
  if (/text\/event-stream|\bServerSentEvent\b|\bcreateSSE\b/.test(source)) return "sse";
  return "json";
}
