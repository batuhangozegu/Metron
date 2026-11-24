import { test } from "node:test";
import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

// A broken encoding once replaced ı, ş, ğ, ç, ö, ü and İ with "?" in these files.
const dir = join(import.meta.dirname, "../src/messages/tr");
const lostLetter = /[A-Za-zçğıöşüÇĞİÖŞÜ]\?[A-Za-zçğıöşü]|(^|\s)\?[a-zçğıöşü]|\?\?/;

test("Turkish messages keep their non-ASCII letters", () => {
  for (const file of readdirSync(dir).filter(name => name.endsWith(".json"))) {
    for (const [key, value] of Object.entries(JSON.parse(readFileSync(join(dir, file), "utf8")))) {
      assert.doesNotMatch(value, lostLetter, `${file} ${key}: ${value}`);
    }
  }
});
