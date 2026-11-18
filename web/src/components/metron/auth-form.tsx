"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import { ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { apiError, apiRequest, validCode, validPassword, type UiError, type User } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { Card } from "./card";
import { CodeField, FormError } from "./auth-fields";
import { useAuth } from "./auth-provider";
import { useLocale } from "./locale-provider";

export function AuthForm({ register = false }: { register?: boolean }) {
  const t = getMessages("auth", useLocale());
  const router = useRouter();
  const { user, acceptUser } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [pending, setPending] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<UiError | null>(null);
  useEffect(() => { if (user) router.replace("/"); }, [user, router]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    setError(null);
    if (pending && !validCode(code)) { setError({ key: "code.invalid" }); return; }
    if (!pending) {
      const normalized = email.trim();
      if (!/^[^\s@]+@[^\s@]+$/.test(normalized) || new TextEncoder().encode(normalized).length > 254) {
        setError({ key: "email.invalid" }); return;
      }
      if (register && !validPassword(password)) { setError({ key: "password.invalid" }); return; }
      if (!password) { setError({ key: "error.credentials" }); return; }
    }
    setBusy(true);
    try {
      const result = pending
        ? await apiRequest<User>("/auth/login/2fa", { code })
        : await apiRequest<User | { two_factor_required: true }>(register ? "/auth/register" : "/auth/login", { email: email.trim(), password });
      setPassword("");
      setCode("");
      if ("two_factor_required" in result && result.two_factor_required) {
        acceptUser(null);
        setPending(true);
      } else {
        acceptUser(result as User);
        router.replace("/");
      }
    } catch (error) {
      setError(apiError(error, pending ? "code" : register ? "register" : "login"));
      if (pending) setCode("");
    } finally { setBusy(false); }
  }

  return <section className="auth-layout mx-auto w-full max-w-xl py-6">
    <div className="mb-6 flex items-center gap-3 text-muted"><ShieldCheck aria-hidden="true" className="size-5" /><span className="micro-label">{t("eyebrow")}</span></div>
    <h1>{t(pending ? "code.title" : register ? "register" : "signIn")}</h1>
    <p className="mt-3 mb-8 text-muted">{t(pending ? "code.note" : register ? "register.note" : "signIn.note")}</p>
    <Card className="auth-card p-6 sm:p-8">
      <form noValidate onSubmit={submit} className="flex flex-col gap-5" aria-busy={busy}>
        <fieldset disabled={busy} className="flex min-w-0 flex-col gap-5">
          {pending ? <CodeField autoFocus value={code} onChange={event => setCode(event.target.value)} aria-invalid={error?.key === "code.invalid" || error?.key === "error.code"} /> : <>
            <div className="flex flex-col gap-2">
              <label htmlFor="email" className="micro-label text-muted">{t("email")}</label>
              <Input id="email" name="email" type="email" autoComplete="username" required value={email} onChange={event => setEmail(event.target.value)} aria-invalid={error?.key === "email.invalid"} />
            </div>
            <div className="flex flex-col gap-2">
              <label htmlFor="password" className="micro-label text-muted">{t("password")}</label>
              <Input id="password" name="password" type="password" autoComplete={register ? "new-password" : "current-password"} required value={password} onChange={event => setPassword(event.target.value)} aria-describedby={register ? "password-hint" : undefined} aria-invalid={error?.key === "password.invalid"} />
              {register && <p id="password-hint" className="text-small text-subtle">{t("password.hint")}</p>}
            </div>
          </>}
          <FormError error={error} />
          <Button type="submit" className="w-full">{t(busy ? "working" : pending ? "code.verify" : register ? "register" : "signIn")}</Button>
          {pending && <Button type="button" variant="ghost" onClick={() => { setPending(false); setCode(""); setError(null); }}>{t("code.restart")}</Button>}
        </fieldset>
      </form>
      {!pending && <Link href={register ? "/sign-in" : "/register"} className="mt-6 inline-flex min-h-control items-center text-sm text-muted underline underline-offset-4">{t(register ? "signIn.link" : "register.link")}</Link>}
    </Card>
  </section>;
}
