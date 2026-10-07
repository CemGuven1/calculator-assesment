import type { Operation } from '../api/types.ts'

export interface OperationInfo {
  label: string
  symbol: string
  /** The number of operands, which must match the backend. */
  arity: 1 | 2
}

/** Display details for every operation, in the order the UI shows them. */
export const OPERATIONS: Record<Operation, OperationInfo> = {
  add: { label: 'Add', symbol: '+', arity: 2 },
  subtract: { label: 'Subtract', symbol: '−', arity: 2 },
  multiply: { label: 'Multiply', symbol: '×', arity: 2 },
  divide: { label: 'Divide', symbol: '÷', arity: 2 },
  power: { label: 'Power', symbol: 'xʸ', arity: 2 },
  sqrt: { label: 'Square root', symbol: '√', arity: 1 },
  percentage: { label: 'Percentage', symbol: '%', arity: 2 },
}
