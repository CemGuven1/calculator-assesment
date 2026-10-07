import type { Operation } from '../api/types.ts'
import { OPERATIONS } from './operations.ts'
import { parseOperand } from './parseOperand.ts'

/** The two input fields: the first and second operand. */
export type Field = 'a' | 'b'

/** A finished calculation, kept so the UI can show the full expression. */
export interface Calculation {
  operation: Operation
  operands: number[]
  result: number
}

/** Where the latest request stands. Only one of these can be true at a time. */
export type Status =
  | { kind: 'idle' }
  | { kind: 'loading'; request: { operation: Operation; operands: number[] } }
  | { kind: 'success'; calculation: Calculation }
  | { kind: 'error'; message: string }

export interface CalculatorState {
  operation: Operation
  /** The raw text of each input field. */
  input: Record<Field, string>
  /** Validation messages from the last submit, per field. */
  fieldErrors: Partial<Record<Field, string>>
  status: Status
}

export type CalculatorAction =
  | { type: 'operationSelected'; operation: Operation }
  | { type: 'inputChanged'; field: Field; value: string }
  | { type: 'submitted' }
  | { type: 'requestSucceeded'; result: number }
  | { type: 'requestFailed'; message: string }

export const initialState: CalculatorState = {
  operation: 'add',
  input: { a: '', b: '' },
  fieldErrors: {},
  status: { kind: 'idle' },
}

/**
 * The calculator's state transitions. It is pure: submitting only moves the
 * state to "loading" with the parsed request, and useCalculator sends it.
 *
 * Errors describe the last attempt, so editing an input or picking another
 * operation clears them. A successful result stays visible until the next one
 * replaces it, since it shows its own expression.
 */
export function calculatorReducer(state: CalculatorState, action: CalculatorAction): CalculatorState {
  switch (action.type) {
    case 'operationSelected':
      return {
        ...state,
        operation: action.operation,
        fieldErrors: {},
        status: clearError(state.status),
      }

    case 'inputChanged':
      return {
        ...state,
        input: { ...state.input, [action.field]: action.value },
        fieldErrors: { ...state.fieldErrors, [action.field]: undefined },
        status: clearError(state.status),
      }

    case 'submitted': {
      if (state.status.kind === 'loading') {
        return state
      }
      const fields: Field[] = OPERATIONS[state.operation].arity === 1 ? ['a'] : ['a', 'b']
      const operands: number[] = []
      const fieldErrors: Partial<Record<Field, string>> = {}
      for (const field of fields) {
        const parsed = parseOperand(state.input[field])
        if (parsed.ok) {
          operands.push(parsed.value)
        } else {
          fieldErrors[field] = parsed.error
        }
      }
      if (Object.keys(fieldErrors).length > 0) {
        return { ...state, fieldErrors, status: clearError(state.status) }
      }
      return {
        ...state,
        fieldErrors: {},
        status: { kind: 'loading', request: { operation: state.operation, operands } },
      }
    }

    // Responses only count while their request is in flight.
    case 'requestSucceeded':
      if (state.status.kind !== 'loading') {
        return state
      }
      return {
        ...state,
        status: { kind: 'success', calculation: { ...state.status.request, result: action.result } },
      }

    case 'requestFailed':
      if (state.status.kind !== 'loading') {
        return state
      }
      return { ...state, status: { kind: 'error', message: action.message } }
  }
}

function clearError(status: Status): Status {
  return status.kind === 'error' ? { kind: 'idle' } : status
}
