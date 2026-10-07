export type ParseResult = { ok: true; value: number } | { ok: false; error: string }

// An optional sign, digits with an optional decimal point (or a leading
// point, as in .5), and an optional exponent. Unlike Number(), this rejects
// hex, "Infinity" and the empty string.
const NUMBER = /^[+-]?(\d+\.?\d*|\.\d+)(e[+-]?\d+)?$/i
const DECIMAL_COMMA = /^[+-]?\d+,\d+$/

/**
 * Parses the text of an input field into a finite number, or explains what is
 * wrong with it. It only checks that the text is a number; rules such as
 * "no division by zero" belong to the backend.
 */
export function parseOperand(text: string): ParseResult {
  const trimmed = text.trim()
  if (trimmed === '') {
    return { ok: false, error: 'Enter a number.' }
  }
  if (DECIMAL_COMMA.test(trimmed)) {
    return { ok: false, error: 'Use a dot for decimals, like 1.5.' }
  }
  if (!NUMBER.test(trimmed)) {
    return { ok: false, error: 'Enter a valid number, like -3.5 or 1e3.' }
  }
  const value = Number(trimmed)
  if (!Number.isFinite(value)) {
    return { ok: false, error: 'This number is too large.' }
  }
  return { ok: true, value }
}
