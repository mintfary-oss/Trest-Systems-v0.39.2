/** One API namespace in the browser, Docker DNS only during server rendering. */
const API_BASE = typeof window !== "undefined"
  ? (process.env.NEXT_PUBLIC_API_URL || "/marketplace-api")
  : (process.env.INTERNAL_API_URL || "http://marketplace-api:8000");
export async function apiFetch(path, options = {}) {
  const {headers = {}, ...rest} = options;
  const res = await fetch(`${API_BASE}${path}`, {
    ...rest,
    headers: {"Content-Type": "application/json", ...headers},
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({detail: res.statusText}));
    throw new Error(typeof err.detail === "string" ? err.detail : `API error ${res.status}`);
  }
  return res.status === 204 ? null : res.json();
}
export async function adminFetch(path, token, options = {}) {
  return apiFetch(path, {...options, headers: {...options.headers, Authorization: `Bearer ${token}`}});
}
