import { OPERATIONS } from './operations.ts'
import type { Calculation } from './reducer.ts'

/**
 * Formats a number from the backend for display. It changes only how the
 * result looks, never the result itself:
 *
 * - Whole numbers up to 2^53 are exact in float64, so every digit is shown.
 * - Other values are rounded to 15 significant digits, which hides float64
 *   noise (0.30000000000000004 shows as 0.3).
 * - From 10^15 up, those use exponent form. Written out in full, the rounded
 *   value would end in zeros that look exact but aren't (2^60 would show as
 *   1152921504606850000 instead of 1152921504606846976).
 */
export function formatNumber(value: number): string {
  if (Number.isSafeInteger(value)) {
    return String(value) // also turns -0 into "0"
  }
  const rounded = Number(value.toPrecision(15))
  if (!Number.isFinite(rounded)) {
    // Rounding the largest float64 values up overflows, so show them unrounded.
    return String(value)
  }
  return Math.abs(rounded) >= 1e15 ? rounded.toExponential() : String(rounded)
}

/** Like formatNumber, but negative operands get parentheses: 5 − (-3). */
function formatOperand(value: number): string {
  const text = formatNumber(value)
  return value < 0 ? `(${text})` : text
}

/** Splits a calculation into the expression and result to display. */
export function describeCalculation({ operation, operands, result }: Calculation): {
  expression: string
  result: string
} {
  const [a, b] = operands.map(formatOperand)
  let expression: string
  switch (operation) {
    case 'sqrt':
      expression = `√${a}`
      break
    case 'percentage':
      expression = `${a}% of ${b}`
      break
    case 'power':
      expression = `${a} ^ ${b}`
      break
    default:
      expression = `${a} ${OPERATIONS[operation].symbol} ${b}`
  }
  return { expression, result: formatNumber(result) }
}
