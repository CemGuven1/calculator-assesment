export type ParseResult = { ok: true; value: number } | { ok: false; error: string }

// An optional sign, digits with an optional decimal point (or a leading
// point, as in .5), and an optional exponent. Unlike Number(), this rejects
// hex, "Infinity" and the empty string.
const NUMBER = /^[+-]?(\d+\.?\d*|\.\d+)(e[+-]?\d+)?$/i
// A comma or space inside a number: a decimal comma (1,5) or a thousands
// separator (1,000 or 1 000). The hint covers both, so that following it
// never turns 1,000 into 1.000, which is 1.
const SEPARATOR = /[,\s]/
const SECOND_POINT = /\..*\./

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
  if (SEPARATOR.test(trimmed)) {
    return { ok: false, error: 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.' }
  }
  if (SECOND_POINT.test(trimmed)) {
    return { ok: false, error: 'A number can only have one decimal point.' }
  }
  if (!NUMBER.test(trimmed)) {
    return { ok: false, error: 'Enter a valid number, like -3.5 or 1e3.' }
  }
  const value = Number(trimmed)
  if (!Number.isFinite(value)) {
    return { ok: false, error: 'This number is too large.' }
  }
  // Number() turns a non-zero number too close to zero for float64 into 0.
  const mantissa = trimmed.toLowerCase().split('e')[0]
  if (value === 0 && /[1-9]/.test(mantissa)) {
    return { ok: false, error: 'This number is too small.' }
  }
  return { ok: true, value }
}
