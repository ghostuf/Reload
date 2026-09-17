import { API_BASE_URL, ApiError } from "@/lib/api";
import type { User, UserResponse } from "@/types/user";

/**
 * The API stores the session in HttpOnly cookies on its own origin, so every
 * call here must run in the browser: a Server Action would receive the
 * Set-Cookie header on the Next.js server, where it does the user no good.
 * `credentials: "include"` is what attaches those cookies to a cross-origin
 * request, and the API echoes a single allowed origin so they are accepted.
 */

/** Error code the API returns with a 401 when only the access token expired. */
const TOKEN_EXPIRED = "token_expired";

type ErrorBody = {
  error?: string;
  code?: string;
};

/** readError pulls the API's message out of a failed response. */
async function readError(res: Response, fallback: string): Promise<ApiError> {
  const body: ErrorBody | null = await res.json().catch(() => null);
  return new ApiError(body?.error || fallback, res.status);
}

/**
 * refreshInFlight collapses concurrent refreshes into one request. Without it,
 * two calls failing at once would both rotate, and the second would present a
 * token the first had already spent — which the API treats as a replay and
 * answers by revoking the whole session.
 */
let refreshInFlight: Promise<boolean> | null = null;

/** refreshSession rotates the token pair, returning false once the session is over. */
export function refreshSession(): Promise<boolean> {
  refreshInFlight ??= (async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      return res.ok;
    } catch {
      // Offline or the API is down — indistinguishable from here, and neither
      // means the session is invalid.
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

/**
 * authedFetch sends a credentialed request and, if the access token has merely
 * expired, refreshes once and retries. Any other 401 is returned as-is: it means
 * the session itself is gone, and retrying would only produce a second failure.
 */
export async function authedFetch(
  path: string,
  init: RequestInit = {},
): Promise<Response> {
  const request: RequestInit = { ...init, credentials: "include" };

  const res = await fetch(`${API_BASE_URL}${path}`, request);
  if (res.status !== 401) return res;

  // Read from a clone so the caller still gets an unconsumed body if the
  // response turns out not to be refreshable.
  const body: ErrorBody | null = await res.clone().json().catch(() => null);
  if (body?.code !== TOKEN_EXPIRED) return res;

  if (!(await refreshSession())) return res;

  return fetch(`${API_BASE_URL}${path}`, request);
}

function jsonRequest(body: unknown): RequestInit {
  return {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  };
}

export type RegisterParams = {
  email: string;
  password: string;
  displayName?: string;
};

/**
 * register creates the account. It does not sign the user in — the API returns
 * the new user without setting cookies — so callers follow it with login.
 */
export async function register(params: RegisterParams): Promise<User> {
  const res = await fetch(
    `${API_BASE_URL}/auth/register`,
    jsonRequest({
      email: params.email,
      password: params.password,
      display_name: params.displayName ?? "",
    }),
  );

  if (!res.ok) {
    throw await readError(res, "Could not create your account");
  }

  const body: UserResponse = await res.json();
  return body.data;
}

export async function login(email: string, password: string): Promise<User> {
  const res = await fetch(`${API_BASE_URL}/auth/login`, {
    ...jsonRequest({ email, password }),
    credentials: "include",
  });

  if (!res.ok) {
    throw await readError(res, "Could not sign you in");
  }

  const body: UserResponse = await res.json();
  return body.data;
}

/**
 * logout revokes the refresh token server-side. The API clears the cookies on
 * its own, so a failure here still leaves the client signed out locally.
 */
export async function logout(): Promise<void> {
  await fetch(`${API_BASE_URL}/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

/**
 * getCurrentUser returns the signed-in user, or null when there is no session.
 * A missing session is an ordinary state, not an error, so only unexpected
 * failures throw.
 */
export async function getCurrentUser(): Promise<User | null> {
  const res = await authedFetch("/auth/me");

  if (res.status === 401) return null;
  if (!res.ok) {
    throw await readError(res, "Could not load your account");
  }

  const body: UserResponse = await res.json();
  return body.data;
}
