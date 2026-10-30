import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AuthGate, AuthProvider } from "@/app/auth";
import { createMemoryTokenStorage } from "@/shared/api/tokenStorage";

/**
 * Renders a gate over a marker, so the test can tell "app is visible" from "gate
 * is visible" without knowing anything about routing.
 */
const renderGate = (
  authRequired: boolean,
  storage = createMemoryTokenStorage(),
): { readonly storage: ReturnType<typeof createMemoryTokenStorage> } => {
  render(
    <AuthProvider required={authRequired} storage={storage}>
      <AuthGate>
        <main>app content</main>
      </AuthGate>
    </AuthProvider>,
  );

  return { storage };
};

describe("AuthGate", () => {
  it("shows the app when no token is needed", () => {
    renderGate(false);

    expect(screen.getByText("app content")).toBeInTheDocument();
    expect(screen.queryByLabelText("Access token")).not.toBeInTheDocument();
  });

  it("asks for a token when one is required and none is stored", () => {
    renderGate(true);

    expect(screen.getByLabelText("Access token")).toBeInTheDocument();
    expect(screen.queryByText("app content")).not.toBeInTheDocument();
  });

  it("shows the app when a token was stored in an earlier session", () => {
    // This is what makes the token survive a page reload: the gate reads storage
    // once on mount.
    renderGate(true, createMemoryTokenStorage("stored-token"));

    expect(screen.getByText("app content")).toBeInTheDocument();
  });

  it("keeps the app hidden until Save is pressed", async () => {
    const user = userEvent.setup();
    renderGate(true);

    await user.type(screen.getByLabelText("Access token"), "typed");

    // Typing alone must not unlock anything: the value is only a candidate.
    expect(screen.queryByText("app content")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(screen.getByText("app content")).toBeInTheDocument();
    });
  });

  it("persists the saved token through the storage it was given", async () => {
    const user = userEvent.setup();
    const { storage } = renderGate(true);

    await user.type(screen.getByLabelText("Access token"), "  saved-token  ");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(storage.get()).toBe("saved-token");
    });
  });

  it("cannot be submitted while empty or whitespace-only", async () => {
    const user = userEvent.setup();
    renderGate(true);

    const save = screen.getByRole("button", { name: "Save" });
    expect(save).toBeDisabled();

    await user.type(screen.getByLabelText("Access token"), "   ");
    expect(save).toBeDisabled();

    await user.type(screen.getByLabelText("Access token"), "x");
    expect(save).toBeEnabled();
  });

  it("submits on Enter as well as on the button", async () => {
    const user = userEvent.setup();
    renderGate(true);

    await user.type(screen.getByLabelText("Access token"), "typed{Enter}");

    await waitFor(() => {
      expect(screen.getByText("app content")).toBeInTheDocument();
    });
  });
});
