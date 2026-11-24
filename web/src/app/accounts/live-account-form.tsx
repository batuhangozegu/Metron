"use client";

import { useRef, useState, type FormEvent, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError, apiRequest, type TradingAccount } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { tradingError } from "@/lib/trading";

/** Two or more mutually exclusive options drawn as one bordered strip of buttons. */
export function Segmented<T extends string>({ name, legend, value, options, onChange }: {
  name: string; legend: ReactNode; value: T; options: readonly { value: T; label: ReactNode }[]; onChange: (value: T) => void;
}) {
  return <fieldset className="flex min-w-0 flex-col gap-2">
    <legend className="micro-label mb-2 text-muted">{legend}</legend>
    <div className="grid gap-1 border border-line bg-raised p-1" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}>
      {options.map(option => <label key={option.value} className={`relative flex min-h-control cursor-pointer items-center justify-center border px-3 py-2 text-sm focus-within:outline-2 focus-within:outline-offset-2 ${value === option.value ? "border-line bg-surface text-text" : "border-transparent text-muted"}`}>
        <input type="radio" name={name} value={option.value} checked={value === option.value} onChange={() => onChange(option.value)} className="sr-only" />
        {option.label}
      </label>)}
    </div>
  </fieldset>;
}

export function LiveAccountForm({ onCreated }: { onCreated: (account: TradingAccount) => void }) {
  const locale = useLocale();
  const t = getMessages("accounts", locale);
  const auth = getMessages("auth", locale);
  const { refresh } = useAuth();
  const pending = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState(false);
  const [market, setMarket] = useState<"spot" | "futures">("spot");
  const [environment, setEnvironment] = useState<"testnet" | "mainnet">("testnet");
  async function connect(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (pending.current) return;
    const form = event.currentTarget;
    const fields = new FormData(form);
    const name = String(fields.get("name") ?? "").trim();
    const api_key = String(fields.get("api_key") ?? "").trim();
    const api_secret = String(fields.get("api_secret") ?? "");
    setCreated(false);
    if (!name || Array.from(name).length > 100 || !api_key || !api_secret.trim()) { setError(t("live.invalid")); return; }
    pending.current = true;
    setBusy(true);
    setError(null);
    try {
      const account = await apiRequest<TradingAccount>("/accounts/live", { name, market, environment, api_key, api_secret });
      form.reset();
      onCreated(account);
      setCreated(true);
    } catch (error) {
      setError(error instanceof ApiError && error.status === 422 ? t("live.keys") : tradingError(error, locale));
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally {
      // Never retain credentials after an attempt or copy them to client storage.
      form.querySelectorAll<HTMLInputElement>('input[name="api_key"], input[name="api_secret"]').forEach(input => { input.value = ""; });
      pending.current = false;
      setBusy(false);
    }
  }
  return <form onSubmit={connect} className="p-6" aria-busy={busy}>
    <fieldset disabled={busy} className="flex min-w-0 flex-col gap-5">
      <div className="flex flex-col gap-2"><label htmlFor="live-name" className="micro-label text-muted">{t("name")}</label><Input id="live-name" name="name" required maxLength={100} /></div>
      <Segmented name="live-market" legend={t("market")} value={market} onChange={setMarket}
        options={[{ value: "spot", label: t("market.spot") }, { value: "futures", label: t("market.futures") }]} />
      <Segmented name="live-environment" legend={t("environment")} value={environment} onChange={setEnvironment}
        options={[{ value: "testnet", label: t("testnet") }, { value: "mainnet", label: <span className="text-limit">{t("mainnet")}</span> }]} />
      <div className="flex flex-col gap-2"><label htmlFor="live-key" className="micro-label text-muted">{t("apiKey")}</label><Input id="live-key" name="api_key" required autoComplete="off" spellCheck={false} /></div>
      <div className="flex flex-col gap-2"><label htmlFor="live-secret" className="micro-label text-muted">{t("apiSecret")}</label><Input id="live-secret" name="api_secret" type="password" required autoComplete="new-password" /></div>
      <p className="text-small text-caution">{t("safety")}</p>
      {error && <p role="alert" className="text-small text-limit">{error}</p>}
      {created && <p role="status" className="text-small">{t("live.created")}</p>}
      <Button type="submit" className="self-start">{busy ? auth("working") : t("connect")}</Button>
    </fieldset>
  </form>;
}
