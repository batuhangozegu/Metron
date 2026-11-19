"use client";

import { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/components/metron/auth-provider";
import { ApiError, apiRequest, type TradingAccount } from "@/lib/auth";

export function usePaperAccounts(initialId = "") {
  const { refresh } = useAuth();
  const [accounts, setAccounts] = useState<TradingAccount[] | null>(null);
  const [selectedId, setSelectedId] = useState(initialId);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback((signal?: AbortSignal) => {
    return apiRequest<TradingAccount[]>("/accounts", undefined, signal).then(result => {
      if (signal?.aborted) return;
      setAccounts(result.filter(account => account.type === "paper" && account.quote_asset === "USDT"));
      setError(null);
    }, error => {
      if (signal?.aborted) return;
      setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
  }, [refresh]);
  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);
  const account = accounts?.find(account => account.id === selectedId) ?? accounts?.[0];
  return { accounts, account, setSelectedId, error, reload: load };
}
