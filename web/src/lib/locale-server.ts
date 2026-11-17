import "server-only";
import { cookies } from "next/headers";
import { localeCookie, parseLocale } from "./messages";

export async function getLocale() {
  return parseLocale((await cookies()).get(localeCookie)?.value);
}
