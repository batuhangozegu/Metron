"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useAuth } from "@/components/metron/auth-provider";
import { ApiError } from "@/lib/auth";
import { getJournal, getPortfolio, type Journal, type Portfolio, type PortfolioPeriod } from "@/lib/portfolio";

export function usePortfolio(accountId: string, period: PortfolioPeriod) {
  const { refresh } = useAuth();
  const [revision, setRevision] = useState(0);
  const key = `${accountId}:${period}:${revision}`;
  const [result, setResult] = useState<{ key: string; data: Portfolio | null; error: unknown } | null>(null);
  const reload = useCallback(() => setRevision(value => value + 1), []);
  useEffect(() => {
    const controller = new AbortController();
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]);
    void getPortfolio(accountId, period, signal).then(data => {
      if (!controller.signal.aborted) setResult({ key, data, error: null });
    }, error => {
      if (controller.signal.aborted) return;
      setResult({ key, data: null, error });
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
    return () => controller.abort();
  }, [accountId, period, key, refresh]);
  return { data: result?.key === key ? result.data : null, error: result?.key === key ? result.error : null, reload };
}

export function useJournal(accountId: string) {
  const { refresh } = useAuth();
  const [revision, setRevision] = useState(0);
  const key = `${accountId}:${revision}`;
  const [result, setResult] = useState<{ key: string; data: Journal | null; error: unknown } | null>(null);
  const [more, setMore] = useState<{ key: string; loading: boolean; error: unknown } | null>(null);
  const request = useRef<AbortController | null>(null);
  const reload = useCallback(() => { request.current?.abort(); setRevision(value => value + 1); }, []);
  useEffect(() => {
    const controller = new AbortController();
    void getJournal(accountId, undefined, AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)])).then(data => {
      if (!controller.signal.aborted) setResult({ key, data, error: null });
    }, error => {
      if (controller.signal.aborted) return;
      setResult({ key, data: null, error });
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
    return () => { controller.abort(); request.current?.abort(); request.current = null; };
  }, [accountId, key, refresh]);
  const data = result?.key === key ? result.data : null;
  const loadMore = async () => {
    if (!data?.next_before || request.current) return;
    const controller = new AbortController();
    request.current = controller;
    setMore({ key, loading: true, error: null });
    try {
      const page = await getJournal(accountId, data.next_before, AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]));
      if (controller.signal.aborted) return;
      const seen = new Set(data.trades.map(trade => trade.position_id));
      setResult({ key, error: null, data: { trades: [...data.trades, ...page.trades.filter(trade => !seen.has(trade.position_id))], next_before: page.next_before } });
      setMore({ key, loading: false, error: null });
    } catch (error) {
      if (controller.signal.aborted) return;
      setMore({ key, loading: false, error });
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { if (request.current === controller) request.current = null; }
  };
  return { data, error: result?.key === key ? result.error : null, reload, loadMore,
    loadingMore: more?.key === key && more.loading, moreError: more?.key === key ? more.error : null };
}
