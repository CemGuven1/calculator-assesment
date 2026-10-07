import { OPERATIONS } from './operations.ts'
import type { Calculation } from './reducer.ts'

/**
 * Formats a number from the backend for display. Rounding to 15 significant
 * digits hides float64 noise (0.30000000000000004 shows as 0.3); it changes
 * only how the backend's result looks, not the result itself.
 */
export function formatNumber(value: number): string {
  const rounded = Number(value.toPrecision(15))
  // Rounding the largest float64 values up overflows to Infinity, so those
  // are shown unrounded. String() also turns -0 into "0".
  return String(Number.isFinite(rounded) ? rounded : value)
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
