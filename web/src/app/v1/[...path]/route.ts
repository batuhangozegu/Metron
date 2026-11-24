import type { NextRequest } from "next/server";

type Context = { params: Promise<{ path: string[] }> };

/** Runtime configuration is intentional: Compose sets the API URL when starting web. */
async function proxy(request: NextRequest, context: Context) {
  const { path } = await context.params;
  const base = (process.env.METRON_API_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
  const headers = new Headers();
  for (const name of ["cookie", "content-type", "user-agent"]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  // Preserve the submitted content type; never turn a cross-site form into JSON.
  const body = request.method === "POST" || request.method === "PUT" ? await request.text() : undefined;
  if (body && new TextEncoder().encode(body).length > 4096) {
    return Response.json({ error: "request too large" }, { status: 400, headers: { "Cache-Control": "no-store" } });
  }
  try {
    const upstream = await fetch(`${base}/v1/${path.map(encodeURIComponent).join("/")}${request.nextUrl.search}`, {
      method: request.method,
      headers,
      body,
      cache: "no-store",
      redirect: "manual",
      signal: AbortSignal.timeout(10_000),
    });
    const responseHeaders = new Headers({ "Cache-Control": "no-store" });
    for (const name of ["content-type", "retry-after"]) {
      const value = upstream.headers.get(name);
      if (value) responseHeaders.set(name, value);
    }
    // Login/logout can set both full and pending cookies. Keep separate Set-Cookie headers.
    for (const cookie of upstream.headers.getSetCookie()) responseHeaders.append("Set-Cookie", cookie);
    return new Response(upstream.body, { status: upstream.status, headers: responseHeaders });
  } catch {
    return Response.json({ error: "service unavailable" }, { status: 503, headers: { "Cache-Control": "no-store" } });
  }
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const DELETE = proxy;
