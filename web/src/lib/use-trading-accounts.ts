"use client";

import { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/components/metron/auth-provider";
import { selectTradingAccount, tradableAccounts } from "@/lib/accounts";
import { ApiError, apiRequest, type TradingAccount } from "@/lib/auth";

export function useTradingAccounts(initialId = "") {
  const { refresh } = useAuth();
  const [accounts, setAccounts] = useState<TradingAccount[] | null>(null);
  const [selectedId, setSelectedId] = useState(initialId);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback((signal?: AbortSignal) => {
    return apiRequest<TradingAccount[]>("/accounts", undefined, signal).then(result => {
      if (signal?.aborted) return;
      setAccounts(tradableAccounts(result));
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
  const account = accounts ? selectTradingAccount(accounts, selectedId) : undefined;
  return { accounts, account, setSelectedId, error, reload: load };
}
