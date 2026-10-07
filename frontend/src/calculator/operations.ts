import type { Operation } from '../api/types.ts'

export interface OperationInfo {
  label: string
  symbol: string
  /**
   * The label of each input field. Their count is the number of operands,
   * which must match the backend.
   */
  operandLabels: readonly [string] | readonly [string, string]
}

/** Display details for every operation, in the order the UI shows them. */
export const OPERATIONS: Record<Operation, OperationInfo> = {
  add: { label: 'Add', symbol: '+', operandLabels: ['First number', 'Second number'] },
  subtract: { label: 'Subtract', symbol: '−', operandLabels: ['First number', 'Second number'] },
  multiply: { label: 'Multiply', symbol: '×', operandLabels: ['First number', 'Second number'] },
  divide: { label: 'Divide', symbol: '÷', operandLabels: ['First number', 'Second number'] },
  power: { label: 'Power', symbol: 'xʸ', operandLabels: ['Base', 'Exponent'] },
  sqrt: { label: 'Square root', symbol: '√', operandLabels: ['Number'] },
  percentage: { label: 'Percent', symbol: '%', operandLabels: ['Percentage (%)', 'Of number'] },
}
