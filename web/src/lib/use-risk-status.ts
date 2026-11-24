"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useAuth } from "@/components/metron/auth-provider";
import { ApiError, apiRequest } from "@/lib/auth";
import { getRiskStatus, type RiskStatus } from "@/lib/risk-status";
import type { Position } from "@/lib/trading";

/** Mount per account, as Portfolio does, to discard snapshots and requests on selection. */
export function useRiskStatus(accountId: string, withPositions = false) {
  const { refresh } = useAuth();
  const [data, setData] = useState<RiskStatus | null>(null);
  const [positions, setPositions] = useState<Position[] | null>(null);
  const [error, setError] = useState<unknown>(null);
  const request = useRef<AbortController | null>(null);
  const load = useCallback(async (replace = false) => {
    if (request.current && !replace) return;
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]);
    try {
      const [status, rows] = await Promise.all([
        getRiskStatus(accountId, signal),
        withPositions ? apiRequest<{ positions: Position[] }>(`/accounts/${encodeURIComponent(accountId)}/positions?status=open`, undefined, signal) : Promise.resolve(null),
      ]);
      if (controller.signal.aborted) return;
      setData(status);
      setPositions(rows?.positions ?? null);
      setError(null);
    } catch (error) {
      if (controller.signal.aborted) return;
      setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { if (request.current === controller) request.current = null; }
  }, [accountId, refresh, withPositions]);
  const reload = useCallback(() => load(true), [load]);
  useEffect(() => {
    const visibleRefresh = () => { if (document.visibilityState === "visible") void load(); };
    visibleRefresh();
    const timer = window.setInterval(visibleRefresh, 5_000);
    document.addEventListener("visibilitychange", visibleRefresh);
    return () => {
      request.current?.abort();
      request.current = null;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", visibleRefresh);
    };
  }, [load]);
  return { data, positions, error, reload };
}
