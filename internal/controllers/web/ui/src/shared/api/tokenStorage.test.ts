import { describe, expect, it } from "vitest";

import {
  createMemoryTokenStorage,
  createTokenStorage,
} from "@/shared/api/tokenStorage";

/**
 * A minimal `Storage` over a `Map`. It is passed in explicitly, so these tests
 * never reach `window.localStorage` and cannot leak state into each other.
 */
const createFakeStorage = (): Storage & { readonly dump: () => Record<string, string> } => {
  const data = new Map<string, string>();

  return {
    get length() {
      return data.size;
    },
    clear: () => {
      data.clear();
    },
    getItem: (key) => data.get(key) ?? null,
    key: (index) => [...data.keys()][index] ?? null,
    removeItem: (key) => {
      data.delete(key);
    },
    setItem: (key, value) => {
      data.set(key, value);
    },
    dump: () => Object.fromEntries(data),
  };
};

describe("createTokenStorage", () => {
  it("starts empty", () => {
    expect(createTokenStorage(createFakeStorage()).get()).toBeUndefined();
  });

  it("round-trips a saved token", () => {
    const storage = createTokenStorage(createFakeStorage());

    storage.set("secret");

    expect(storage.get()).toBe("secret");
  });

  it("trims what it stores and what it reads", () => {
    const backing = createFakeStorage();
    const storage = createTokenStorage(backing);

    storage.set("  secret  ");
    expect(storage.get()).toBe("secret");

    backing.setItem("aimem.token", "  padded  ");
    expect(storage.get()).toBe("padded");
  });

  it("treats a blank token as absent rather than as a credential", () => {
    const storage = createTokenStorage(createFakeStorage());

    storage.set("   ");

    // An empty bearer token is not a token: sending one would produce a 401 that
    // looks like a server problem instead of "the user never entered anything".
    expect(storage.get()).toBeUndefined();
  });

  it("persists under one fixed key", () => {
    const backing = createFakeStorage();

    createTokenStorage(backing).set("secret");

    expect(backing.dump()).toEqual({ "aimem.token": "secret" });
  });
});

describe("createMemoryTokenStorage", () => {
  it("needs no storage implementation", () => {
    const storage = createMemoryTokenStorage();

    storage.set("secret");

    expect(storage.get()).toBe("secret");
    expect(storage.current()).toBe("secret");
  });

  it("starts from an initial value", () => {
    expect(createMemoryTokenStorage(" preseeded ").get()).toBe("preseeded");
  });

  it("normalises a blank value the same way the real storage does", () => {
    const storage = createMemoryTokenStorage("secret");

    storage.set("  ");

    expect(storage.get()).toBeUndefined();
  });
});
