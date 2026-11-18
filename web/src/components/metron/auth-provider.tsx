"use client";

import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import { ApiError, apiError, apiRequest, type UiError, type User } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { useLocale } from "./locale-provider";
import { FormError } from "./auth-fields";
import { Button } from "@/components/ui/button";

type Session = { user: User | null; status: "loading" | "authenticated" | "anonymous" | "error"; error: UiError | null };
interface AuthContextValue extends Session {
  acceptUser: (user: User | null) => void;
  refresh: () => Promise<void>;
}
const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const [session, setSession] = useState<Session>({ user: null, status: "loading", error: null });
  const generation = useRef(0);
  const acceptUser = useCallback((user: User | null) => {
    generation.current++;
    setSession({ user, status: user ? "authenticated" : "anonymous", error: null });
  }, []);
  const refresh = useCallback((signal?: AbortSignal) => {
    const current = ++generation.current;
    return apiRequest<User>("/me", undefined, signal).then(user => {
      if (current === generation.current) setSession({ user, status: "authenticated", error: null });
    }, error => {
      if (signal?.aborted || current !== generation.current) return;
      setSession(error instanceof ApiError && error.status === 401
        ? { user: null, status: "anonymous", error: null }
        : { user: null, status: "error", error: apiError(error) });
    });
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    const onFocus = () => { void refresh(); };
    window.addEventListener("focus", onFocus);
    return () => {
      controller.abort();
      window.removeEventListener("focus", onFocus);
    };
  }, [pathname, refresh]);

  return <AuthContext value={{ ...session, acceptUser, refresh }}>{children}</AuthContext>;
}

export function useAuth() {
  const auth = useContext(AuthContext);
  if (!auth) throw new Error("useAuth requires AuthProvider");
  return auth;
}

/** UI gate only. Every private data request is also authorized by the Go API. */
export function AuthGate({ children }: { children: ReactNode }) {
  const { status, error, refresh } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const t = getMessages("auth", useLocale());
  const publicPage = pathname === "/markets" || pathname === "/sign-in" || pathname === "/register";
  useEffect(() => {
    if (!publicPage && status === "anonymous") router.replace("/sign-in");
  }, [publicPage, status, router]);
  if (publicPage || status === "authenticated") return children;
  if (status === "error") return <div className="auth-feedback flex flex-col items-start gap-4">
    <FormError error={error} /><Button variant="outline" onClick={() => void refresh()}>{t("retry")}</Button>
  </div>;
  return <p role="status" className="text-muted">{t("loading")}</p>;
}
