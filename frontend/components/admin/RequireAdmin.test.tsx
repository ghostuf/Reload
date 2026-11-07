import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import RequireAdmin from "@/components/admin/RequireAdmin";
import { useAuth } from "@/hooks/useAuth";
import { makeUser } from "@/test/fixtures";
import type { AuthContextValue } from "@/components/AuthProvider";
import type { User } from "@/types/user";

const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace, push: vi.fn(), refresh: vi.fn() }),
  usePathname: () => "/admin/games",
}));
vi.mock("@/hooks/useAuth", () => ({ useAuth: vi.fn() }));

function session(user: User | null, loading = false) {
  vi.mocked(useAuth).mockReturnValue({
    user,
    loading,
    isAdmin: user?.role === "admin",
    signIn: vi.fn(),
    signUp: vi.fn(),
    signOut: vi.fn(),
  } satisfies AuthContextValue);
}

function renderGuarded() {
  render(
    <RequireAdmin>
      <p>Admin dashboard</p>
    </RequireAdmin>,
  );
}

beforeEach(() => replace.mockReset());

describe("admin route guard", () => {
  it("redirects a signed-in non-admin to the homepage without showing the page", async () => {
    session(makeUser({ role: "user" }));
    renderGuarded();

    await waitFor(() => expect(replace).toHaveBeenCalledWith("/"));
    expect(screen.queryByText("Admin dashboard")).not.toBeInTheDocument();
  });

  it("sends a signed-out visitor to log in, returning to the admin page", async () => {
    session(null);
    renderGuarded();

    await waitFor(() =>
      expect(replace).toHaveBeenCalledWith(`/login?next=${encodeURIComponent("/admin/games")}`),
    );
    expect(screen.queryByText("Admin dashboard")).not.toBeInTheDocument();
  });

  it("waits for the session check before deciding", () => {
    session(null, true);
    renderGuarded();

    expect(replace).not.toHaveBeenCalled();
    expect(screen.getByText("Checking your session…")).toBeInTheDocument();
  });

  it("lets an admin through", () => {
    session(makeUser({ role: "admin" }));
    renderGuarded();

    expect(screen.getByText("Admin dashboard")).toBeInTheDocument();
    expect(replace).not.toHaveBeenCalled();
  });
});
