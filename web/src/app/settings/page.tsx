"use client";

import { useState, type FormEvent } from "react";
import { QRCodeSVG } from "qrcode.react";
import { LockKeyhole } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader } from "@/components/metron/card";
import { CodeField, FormError } from "@/components/metron/auth-fields";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { SignOut } from "@/components/metron/session-actions";
import { ApiError, apiError, apiRequest, validCode, type UiError, type User } from "@/lib/auth";
import { getMessages } from "@/lib/messages";

interface Setup { secret: string; otpauth_url: string }

export default function SettingsPage() {
  const { user } = useAuth();
  return user ? <Settings key={user.id} user={user} /> : null;
}

function Settings({ user }: { user: User }) {
  const t = getMessages("auth", useLocale());
  const { acceptUser, refresh } = useAuth();
  const [setup, setSetup] = useState<Setup | null>(null);
  const [code, setCode] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<UiError | null>(null);
  const [notice, setNotice] = useState<"twoFactor.enabledNotice" | "twoFactor.disabledNotice" | null>(null);

  async function change(action: "setup" | "enable" | "disable") {
    if (busy) return;
    setError(null);
    setNotice(null);
    if (action !== "setup" && !validCode(code)) { setError({ key: "code.invalid" }); return; }
    if (action === "disable" && !password) { setError({ key: "error.disable" }); return; }
    setBusy(true);
    try {
      if (action === "setup") {
        setSetup(await apiRequest<Setup>("/auth/2fa/setup", {}));
      } else {
        const result = await apiRequest<{ two_factor_enabled: boolean }>(`/auth/2fa/${action}`, action === "disable" ? { code, password } : { code });
        acceptUser({ ...user, two_factor_enabled: result.two_factor_enabled });
        setSetup(null);
        setNotice(action === "enable" ? "twoFactor.enabledNotice" : "twoFactor.disabledNotice");
      }
    } catch (error) {
      setError(apiError(error, action === "disable" ? "disable" : action === "enable" ? "code" : "settings"));
      // Invalid TOTP also uses 401: check /me before deciding the session is gone.
      if (error instanceof ApiError && (error.status === 401 || error.status === 409)) void refresh();
    } finally {
      setBusy(false);
      setCode("");
      setPassword("");
    }
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void change(user.two_factor_enabled ? "disable" : "enable");
  }

  return <section className="auth-layout flex flex-col gap-6">
    <div><p className="micro-label mb-3 text-subtle">{t("eyebrow")}</p><h1>{t("settings")}</h1><p className="mt-3 text-muted">{t("settings.note")}</p></div>
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
      <Card className="auth-card">
        <CardHeader title={t("profile")} />
        <div className="flex flex-col gap-5 border-t border-line p-6">
          <div><p className="micro-label mb-2 text-subtle">{t("email")}</p><p className="break-all">{user.email}</p></div>
          <SignOut />
        </div>
      </Card>
      <Card className="auth-card">
        <CardHeader title={<span className="flex items-center gap-2"><LockKeyhole aria-hidden="true" className="size-5" />{t("twoFactor")}</span>} note={t("twoFactor.note")} />
        <div className="flex flex-col gap-5 border-t border-line p-6" aria-busy={busy}>
          <p className="micro-label text-text">{t(user.two_factor_enabled ? "twoFactor.enabled" : "twoFactor.disabled")}</p>
          {!user.two_factor_enabled && !setup && <Button className="self-start" disabled={busy} onClick={() => void change("setup")}>{t(busy ? "working" : "twoFactor.setup")}</Button>}
          {!user.two_factor_enabled && setup && <div className="flex flex-col gap-4">
            <p className="text-small text-muted">{t("twoFactor.scan")}</p>
            <QRCodeSVG value={setup.otpauth_url} size={224} level="M" marginSize={4} title={t("twoFactor.qr")} role="img" aria-label={t("twoFactor.qr")} className="auth-qr max-w-full self-start" />
            <div><p className="micro-label mb-2 text-subtle">{t("twoFactor.secret")}</p><code className="num block break-all border border-line bg-raised p-3 text-sm select-all">{setup.secret}</code></div>
            <details><summary className="cursor-pointer text-small text-muted">{t("twoFactor.url")}</summary><code className="num mt-3 block break-all text-xs text-muted select-all">{setup.otpauth_url}</code></details>
          </div>}
          {(user.two_factor_enabled || setup) && <form noValidate onSubmit={submit}>
            <fieldset disabled={busy} className="flex min-w-0 flex-col gap-5">
              {user.two_factor_enabled && <>
                <p className="text-small text-muted">{t("twoFactor.disableNote")}</p>
                <div className="flex flex-col gap-2"><label htmlFor="disable-password" className="micro-label text-muted">{t("password")}</label><Input id="disable-password" name="password" type="password" autoComplete="current-password" value={password} required onChange={event => setPassword(event.target.value)} /></div>
              </>}
              <CodeField id="settings-code" value={code} onChange={event => setCode(event.target.value)} aria-invalid={error?.key === "code.invalid" || error?.key === "error.code"} />
              <Button type="submit" variant="outline" className="self-start whitespace-normal text-left">{t(busy ? "working" : user.two_factor_enabled ? "twoFactor.disable" : "twoFactor.enable")}</Button>
            </fieldset>
          </form>}
          <FormError error={error} />
          {notice && <p role="status" className="border border-line bg-raised p-3 text-small">{t(notice)}</p>}
          <p className="text-small text-subtle">{t("twoFactor.noRecovery")}</p>
        </div>
      </Card>
    </div>
  </section>;
}
