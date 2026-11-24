"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { AccountBadges } from "@/components/metron/account-badges";
import { LiveAccountForm, Segmented } from "./live-account-form";
import { tradingError } from "@/lib/trading";
import { WalletCards } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader } from "@/components/metron/card";
import { FormError } from "@/components/metron/auth-fields";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError, apiError, apiRequest, validPaperBalance, type TradingAccount, type UiError } from "@/lib/auth";
import { formatAmount } from "@/lib/format";
import { getMessages } from "@/lib/messages";

export default function AccountsPage() {
  const locale = useLocale();
  const t = getMessages("accounts", locale);
  const auth = getMessages("auth", locale);
  const { refresh } = useAuth();
  const [accounts, setAccounts] = useState<TradingAccount[] | null>(null);
  const [loadError, setLoadError] = useState<UiError | null>(null);
  const [error, setError] = useState<UiError | null>(null);
  const [invalid, setInvalid] = useState<"name" | "balance" | null>(null);
  const [name, setName] = useState("");
  const [market, setMarket] = useState<"spot" | "futures">("spot");
  const [balance, setBalance] = useState("10000");
  const [busy, setBusy] = useState(false);
  const [created, setCreated] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleted, setDeleted] = useState(false);
  const [mode, setMode] = useState<"paper" | "binance">("paper");

  const load = useCallback((signal?: AbortSignal) => {
    return apiRequest<TradingAccount[]>("/accounts", undefined, signal).then(result => {
      if (!signal?.aborted) { setAccounts(result); setLoadError(null); }
    }, error => {
      if (signal?.aborted) return;
      setLoadError(apiError(error, "accounts"));
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
  }, [refresh, setAccounts, setLoadError]);
  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    setError(null);
    setCreated(false);
    const length = Array.from(name.trim()).length;
    if (length < 1 || length > 100) { setInvalid("name"); return; }
    if (!validPaperBalance(balance)) { setInvalid("balance"); return; }
    setInvalid(null);
    setBusy(true);
    try {
      const account = await apiRequest<TradingAccount>("/accounts", { name: name.trim(), paper_balance: balance, market });
      setAccounts(previous => previous ? [...previous, account] : previous);
      setName("");
      setCreated(true);
      // A failed initial list must not hide pre-existing accounts after creation.
      if (accounts === null) void load();
    } catch (error) {
      setError(apiError(error, "accounts"));
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { setBusy(false); }
  }

  async function deleteAccount(id: string) {
    if (deleting) return;
    setDeleting(id); setDeleteError(null); setDeleted(false);
    try {
      await apiRequest(`/accounts/${encodeURIComponent(id)}`, undefined, undefined, "DELETE");
      setAccounts(previous => previous?.filter(account => account.id !== id) ?? null);
      setDeleted(true);
    } catch (error) {
      setDeleteError(error instanceof ApiError && error.status === 409 ? t("delete.conflict") : tradingError(error, locale));
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { setDeleting(null); }
  }

  return <section className="flex min-w-0 flex-col gap-6">
    <div><p className="micro-label mb-3 flex items-center gap-2 text-subtle"><WalletCards aria-hidden="true" className="size-4" />{t("eyebrow")}</p><h1>{t("title")}</h1><p className="mt-3 text-muted">{t("note")}</p></div>
    <div className="grid items-start gap-6 lg:grid-cols-2">
      <Card>
        <CardHeader title={t("list")} />
        <div className="border-t border-line">
          {deleteError && <p role="alert" className="p-6 text-limit">{deleteError}</p>}
          {deleted && <p role="status" className="p-6 text-muted">{t("delete.done")}</p>}
          {loadError ? <div className="flex flex-col items-start gap-3 p-6"><FormError error={loadError} /><Button variant="outline" onClick={() => void load()}>{auth("retry")}</Button></div>
            : accounts === null ? <p role="status" className="p-6 text-muted">{t("loading")}</p>
            : accounts.length === 0 ? <p className="p-6 text-muted">{t("empty")}</p>
            : <ul className="divide-y divide-line">{accounts.map(account => <li key={account.id} className="flex flex-wrap items-center justify-between gap-4 p-6">
              <div className="min-w-0"><h3 className="break-all text-base font-semibold">{account.name}</h3><p className="micro-label mt-2 flex flex-wrap items-center gap-2 text-subtle"><AccountBadges account={account} locale={locale} /><span className="inline-flex border border-line px-2 py-1">{t(account.market === "futures" ? "market.futures" : "market.spot")}</span></p></div>
              <div className="min-w-0"><p className="micro-label text-subtle">{t("balance.label")}</p><p className="num break-all text-xl">{account.paper_balance === null ? t("balance.unavailable") : formatAmount(account.paper_balance, Math.max(2, account.paper_balance.split(".")[1]?.length ?? 0), locale)} <span className="text-small text-muted">{account.quote_asset}</span></p></div>
              {account.type === "live" && <Button variant="outline" disabled={deleting !== null} onClick={() => void deleteAccount(account.id)}>{deleting === account.id ? auth("working") : t("delete")}</Button>}
            </li>)}</ul>}
        </div>
      </Card>
      <Card>
        <CardHeader title={t("add")} />
        <div className="border-t border-line px-6 pt-6">
          <div role="tablist" className="grid grid-cols-2 gap-1 border border-line bg-raised p-1">
            {(["paper", "binance"] as const).map(value => <button key={value} type="button" role="tab" aria-selected={mode === value} onClick={() => setMode(value)}
              className={`min-h-control border px-3 py-2 text-sm font-semibold ${mode === value ? "border-line bg-surface text-text" : "border-transparent text-muted hover:text-text"}`}>{t(`mode.${value}`)}</button>)}
          </div>
        </div>
        {mode === "binance" ? <LiveAccountForm onCreated={account => { setAccounts(previous => previous ? [...previous, account] : previous); if (accounts === null) void load(); }} />
        : <form noValidate onSubmit={create} className="p-6" aria-busy={busy}>
          <fieldset disabled={busy} className="flex min-w-0 flex-col gap-5">
            <div className="flex flex-col gap-2">
              <label htmlFor="account-name" className="micro-label text-muted">{t("name")}</label>
              <Input id="account-name" name="name" required value={name} onChange={event => setName(event.target.value)} aria-invalid={invalid === "name"} aria-describedby="name-hint" />
              <p id="name-hint" className="text-small text-subtle">{t("name.hint")}</p>
            </div>
            <Segmented name="market" legend={t("market")} value={market} onChange={setMarket}
              options={[{ value: "spot", label: t("market.spot") }, { value: "futures", label: t("market.futures") }]} />
            <div className="flex flex-col gap-2">
              <label htmlFor="paper-balance" className="micro-label text-muted">{t("balance")}</label>
              <Input id="paper-balance" name="paper_balance" type="text" inputMode="decimal" required value={balance} onChange={event => setBalance(event.target.value)} aria-invalid={invalid === "balance"} aria-describedby="balance-hint" />
              <p id="balance-hint" className="text-small text-subtle">{t("balance.hint")}</p>
            </div>
            {invalid && <p role="alert" className="auth-error text-small text-limit">{t(invalid === "name" ? "name.invalid" : "balance.invalid")}</p>}
            <FormError error={error} />
            {created && <p role="status" className="text-small">{t("created")}</p>}
            <Button type="submit" className="self-start">{busy ? auth("working") : t("create")}</Button>
          </fieldset>
        </form>}
      </Card>
    </div>
  </section>;
}
