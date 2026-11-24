"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { Card } from "@/components/metron/card";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { getRiskProfile, putRiskProfile, validRiskCount, validRiskDecimal, type RiskProfile, type RiskProfileInput } from "@/lib/risk";
import { tradingError } from "@/lib/trading";
import { accountLabel } from "@/lib/accounts";
import { AccountBadges } from "@/components/metron/account-badges";
import { useTradingAccounts } from "@/lib/use-trading-accounts";
import { DisciplineMode } from "./discipline-mode";

const control = "num min-h-control w-full min-w-0 rounded-control border border-line bg-raised px-3 text-base text-text";
const fields = [
  { key: "risk_per_trade_percent", label: "riskPerTrade" },
  { key: "total_open_risk_limit_percent", label: "totalOpenRisk", hint: "limits.totalOpenRisk.hint" },
  { key: "max_open_positions", label: "maxPositions", count: true },
  { key: "max_losing_trades_today", label: "losingTradesPerDay", count: true },
  { key: "daily_loss_limit_percent", label: "dailyLoss", hint: "limits.dailyLoss.hint" },
  { key: "max_stopped_trades_this_week", label: "stoppedTradesPerWeek", count: true },
  { key: "weekly_loss_limit_percent", label: "weeklyLoss", hint: "limits.weeklyLoss.hint" },
  { key: "target_risk_reward", label: "targetRiskReward" },
] as const;
type Draft = Record<keyof RiskProfileInput, string>;
const draftOf = (profile: RiskProfile): Draft => Object.fromEntries(fields.map(({ key }) => [key, String(profile[key])])) as Draft;

export function AccountLimits({ initialAccountId }: { initialAccountId?: string }) {
  const locale = useLocale();
  const t = getMessages("rules", locale);
  const { accounts, account, setSelectedId, error, reload } = useTradingAccounts(initialAccountId);
  const router = useRouter();
  const params = useSearchParams();
  const selectAccount = (id: string) => {
    setSelectedId(id);
    const query = new URLSearchParams(params.toString());
    query.set("account", id);
    router.replace(`/rules?${query}`, { scroll: false });
  };

  return <section className="flex min-w-0 flex-col gap-6">
    <div className="flex flex-col gap-2"><h1 className="text-title">{t("title")}</h1><p className="max-w-[620px] text-body text-muted">{t("description")}</p></div>
    {error !== null ? <div className="flex flex-col items-start gap-3"><p role="alert" className="text-limit">{tradingError(error, locale)}</p><Button onClick={() => void reload()}>{t("retry")}</Button></div>
      : accounts === null ? <p role="status" className="text-muted">{t("loadingAccounts")}</p>
      : !account ? <p className="text-muted">{t("noAccounts")} <Link href="/accounts" className="inline-flex min-h-control items-center underline">{t("manageAccounts")}</Link></p>
      : <>
        <div className="flex max-w-md flex-col gap-2"><label htmlFor="rules-account" className="text-small text-muted">{t("account")}</label>
          <select id="rules-account" className={control} value={account.id} onChange={event => selectAccount(event.target.value)}>{accounts.map(item => <option key={item.id} value={item.id}>{accountLabel(item, locale)}</option>)}</select>
          <AccountBadges account={account} locale={locale} />
        </div>
        <ProfileLoader key={account.id} accountId={account.id} />
        <DisciplineMode key={`discipline:${account.id}`} accountId={account.id} locale={locale} />
      </>}
  </section>;
}

function ProfileLoader({ accountId }: { accountId: string }) {
  const locale = useLocale();
  const t = getMessages("rules", locale);
  const { refresh } = useAuth();
  const [profile, setProfile] = useState<RiskProfile | null>(null);
  const [error, setError] = useState<unknown>(null);
  const load = useCallback((signal?: AbortSignal) => {
    return getRiskProfile(accountId, signal).then(result => {
      if (signal?.aborted) return;
      setProfile(result);
      setError(null);
    }, error => {
      if (signal?.aborted) return;
      setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
  }, [accountId, refresh]);
  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);
  if (error !== null) return <div className="flex flex-col items-start gap-3"><p role="alert" className="text-limit">{tradingError(error, locale)}</p><Button onClick={() => void load()}>{t("retry")}</Button></div>;
  if (!profile) return <p role="status" className="text-muted">{t("loadingProfile")}</p>;
  return <ProfileForm accountId={accountId} initialProfile={profile} />;
}

function ProfileForm({ accountId, initialProfile }: { accountId: string; initialProfile: RiskProfile }) {
  const locale = useLocale();
  const t = getMessages("rules", locale);
  const { refresh } = useAuth();
  const [profile, setProfile] = useState(initialProfile);
  const [values, setValues] = useState(() => draftOf(initialProfile));
  const [attempted, setAttempted] = useState(false);
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const pending = useRef(false);
  const lifetime = useRef<AbortController | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    return () => controller.abort();
  }, []);
  const invalid = (field: typeof fields[number]) => !("count" in field ? validRiskCount(values[field.key]) : validRiskDecimal(values[field.key]));
  const dirty = fields.some(({ key }) => values[key] !== String(profile[key]));

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (pending.current) return;
    setAttempted(true);
    setSaved(false);
    setError(null);
    if (fields.some(invalid)) return;
    const body: RiskProfileInput = {
      ...values,
      max_open_positions: Number(values.max_open_positions),
      max_losing_trades_today: Number(values.max_losing_trades_today),
      max_stopped_trades_this_week: Number(values.max_stopped_trades_this_week),
    };
    pending.current = true;
    setBusy(true);
    try {
      const result = await putRiskProfile(accountId, body, AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]));
      if (lifetime.current?.signal.aborted) return;
      setProfile(result);
      setValues(draftOf(result));
      setSaved(true);
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally {
      pending.current = false;
      if (!lifetime.current?.signal.aborted) setBusy(false);
    }
  }

  const updatedAt = profile.updated_at === null ? null : new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" }).format(new Date(profile.updated_at));
  return <form noValidate onSubmit={save} aria-busy={busy} className="flex min-w-0 flex-col gap-4">
    <Card className="flex flex-col gap-5 p-card" aria-labelledby="account-limits-title">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-1"><h2 id="account-limits-title" className="text-heading">{t("limits.title")}</h2><p className="text-small text-subtle">{t("limits.description")}</p></div>
        <p role="status" className="text-small text-muted">{updatedAt === null ? t("notSaved") : t("updatedAt", { time: updatedAt })}</p>
      </div>
      <p className="text-small text-subtle">{t("validationHint")}</p>
      <fieldset disabled={busy} className="grid min-w-0 grid-cols-[repeat(auto-fit,minmax(min(100%,220px),1fr))] gap-x-6 gap-y-5">
        {fields.map(field => {
          const bad = attempted && invalid(field);
          const hint = "hint" in field ? field.hint : null;
          const id = `risk-${field.key}`;
          return <div key={field.key} className="flex min-w-0 flex-col gap-2">
            <label htmlFor={id} className="text-small text-muted">{t(`limits.${field.label}`)}</label>
            <input id={id} name={field.key} inputMode={"count" in field ? "numeric" : "decimal"} value={values[field.key]} aria-invalid={bad} aria-describedby={[hint ? `${id}-hint` : "", bad ? `${id}-error` : ""].filter(Boolean).join(" ") || undefined} onChange={event => { setValues(previous => ({ ...previous, [field.key]: event.target.value })); setSaved(false); setError(null); }} className={control} />
            {hint && <span id={`${id}-hint`} className="text-xs leading-[17px] text-subtle">{t(hint)}</span>}
            {bad && <p id={`${id}-error`} role="alert" className="text-small text-limit">{t("count" in field ? "invalidCount" : "invalidDecimal")}</p>}
          </div>;
        })}
      </fieldset>
    </Card>
    {error !== null && <p role="alert" className="text-limit">{tradingError(error, locale)}</p>}
    <div className="flex flex-wrap items-center justify-end gap-3">
      <span role="status" className="text-small text-muted">{saved ? t("saved") : dirty ? t("unsavedChanges") : ""}</span>
      <Button type="submit" disabled={busy}>{t(busy ? "saving" : "save")}</Button>
    </div>
  </form>;
}
