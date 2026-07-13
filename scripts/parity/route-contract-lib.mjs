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
    .replace(/\[\.\.\.([^\]]+)\]/g, "{$1...}")
    .replace(/\[([^\]]+)\]/g, "{$1}");
}

export function extractRouteMethods(source) {
  const methods = new Set();
  const patterns = [
    /export\s+(?:async\s+)?function\s+([A-Z]+)\b/g,
    /export\s+const\s+([A-Z]+)\b/g,
    /\bas\s+([A-Z]+)\b/g,
  ];

  for (const pattern of patterns) {
    for (const match of source.matchAll(pattern)) {
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
