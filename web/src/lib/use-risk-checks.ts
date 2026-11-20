"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useAuth } from "@/components/metron/auth-provider";
import { ApiError } from "@/lib/auth";
import { getRiskChecks, type RiskCheckHistoryEntry } from "@/lib/risk";

export function useRiskChecks(accountId?: string) {
  const { refresh } = useAuth();
  const generation = useRef(0);
  const [result, setResult] = useState<{ accountId: string; checks: RiskCheckHistoryEntry[] | null; error: unknown } | null>(null);
  const reload = useCallback((signal?: AbortSignal) => {
    if (!accountId) return Promise.resolve();
    const current = ++generation.current;
    return getRiskChecks(accountId, 20, signal).then(data => {
      if (!signal?.aborted && current === generation.current) setResult({ accountId, checks: data.checks, error: null });
    }, error => {
      if (signal?.aborted || current !== generation.current) return;
      setResult({ accountId, checks: null, error });
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
  }, [accountId, refresh]);
  useEffect(() => {
    const controller = new AbortController();
    void reload(controller.signal);
    return () => controller.abort();
  }, [reload]);
  return {
    checks: result?.accountId === accountId ? result?.checks ?? null : null,
    error: result?.accountId === accountId ? result?.error ?? null : null,
    reload,
  };
}
