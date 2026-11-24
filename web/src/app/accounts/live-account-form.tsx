"use client";

import { useRef, useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader } from "@/components/metron/card";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError, apiRequest, type TradingAccount } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { tradingError } from "@/lib/trading";

export function LiveAccountForm({ onCreated }: { onCreated: (account: TradingAccount) => void }) {
  const locale = useLocale();
  const t = getMessages("accounts", locale);
  const auth = getMessages("auth", locale);
  const { refresh } = useAuth();
  const pending = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState(false);
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
      const account = await apiRequest<TradingAccount>("/accounts/live", { name, market: fields.get("market"), environment: fields.get("environment"), api_key, api_secret });
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
  const control = "min-h-control w-full rounded-control border border-line bg-raised px-3 text-text";
  return <Card className="auth-card"><CardHeader title={t("connect")} />
    <form onSubmit={connect} className="border-t border-line p-6" aria-busy={busy}>
      <fieldset disabled={busy} className="flex min-w-0 flex-col gap-5">
        <div className="flex flex-col gap-2"><label htmlFor="live-name" className="micro-label text-muted">{t("name")}</label><Input id="live-name" name="name" required maxLength={100} /></div>
        <div className="flex flex-col gap-2"><label htmlFor="live-market" className="micro-label text-muted">{t("market")}</label><select id="live-market" name="market" defaultValue="spot" className={control}>{(["spot", "futures"] as const).map(market => <option key={market} value={market}>{t(`market.${market}`)}</option>)}</select></div>
        <div className="flex flex-col gap-2"><label htmlFor="live-environment" className="micro-label text-muted">{t("environment")}</label><select id="live-environment" name="environment" defaultValue="testnet" className={control}><option value="testnet">{t("testnet")}</option><option value="mainnet">{t("mainnet")}</option></select></div>
        <div className="flex flex-col gap-2"><label htmlFor="live-key" className="micro-label text-muted">{t("apiKey")}</label><Input id="live-key" name="api_key" required autoComplete="off" spellCheck={false} /></div>
        <div className="flex flex-col gap-2"><label htmlFor="live-secret" className="micro-label text-muted">{t("apiSecret")}</label><Input id="live-secret" name="api_secret" type="password" required autoComplete="new-password" /></div>
        <p className="text-small text-caution">{t("safety")}</p>
        {error && <p role="alert" className="text-small text-limit">{error}</p>}
        {created && <p role="status" className="text-small">{t("live.created")}</p>}
        <Button type="submit" className="self-start">{busy ? auth("working") : t("connect")}</Button>
      </fieldset>
    </form>
  </Card>;
}
