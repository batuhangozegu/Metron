import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

export type CardProps = ComponentProps<"article">;
/** Flat surface with a one-pixel line; add p-card for the padded gauge/card variant. */
export function Card({ className, ...props }: CardProps) {
  return <article className={cn("min-w-0 rounded-card border border-line bg-surface", className)} {...props} />;
}

export interface CardHeaderProps extends Omit<ComponentProps<"div">, "title"> { title: ReactNode; note?: ReactNode; action?: ReactNode }
/** Standard card heading row, with optional explanatory copy and a trailing action. */
export function CardHeader({ title, note, action, className, children, ...props }: CardHeaderProps) {
  return <div className={cn("flex flex-wrap items-center justify-between gap-x-6 gap-y-2 px-6 py-5", className)} {...props}>
    <div><h2 className="text-heading">{title}</h2>{note && <p className="mt-1 text-small text-subtle">{note}</p>}</div>
    {action}{children}
  </div>;
}

export interface StatProps extends ComponentProps<"div"> { label: ReactNode; figure: ReactNode; note?: ReactNode }
/** Compact labeled figure; figures always use tabular mono numerals. */
export function Stat({ label, figure, note, className, ...props }: StatProps) {
  return <div className={cn("flex min-w-0 flex-col gap-1", className)} {...props}>
    <div className="text-small text-subtle">{label}</div>
    <div className="num text-figure-medium">{figure}</div>
    {note && <div className="text-small text-muted">{note}</div>}
  </div>;
}
