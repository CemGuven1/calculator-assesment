import { describe, expect, it } from 'vitest'
import { calculatorReducer, initialState, type CalculatorState, type Status } from './reducer.ts'

function stateWith(overrides: Partial<CalculatorState>): CalculatorState {
  return { ...initialState, ...overrides }
}

const loading: Status = { kind: 'loading', request: { operation: 'add', operands: [2, 3] } }
const success: Status = { kind: 'success', calculation: { operation: 'add', operands: [2, 3], result: 5 } }
const failed: Status = { kind: 'error', message: 'You can’t divide by zero.' }

describe('calculatorReducer', () => {
  it('starts with add, empty inputs and no request', () => {
    expect(initialState).toEqual({
      operation: 'add',
      input: { a: '', b: '' },
      fieldErrors: {},
      status: { kind: 'idle' },
    })
  })

  describe('inputChanged', () => {
    it('updates the input and clears only that field’s error', () => {
      const state = stateWith({ fieldErrors: { a: 'Enter a number.', b: 'Enter a number.' } })

      const next = calculatorReducer(state, { type: 'inputChanged', field: 'a', value: '4' })

      expect(next.input).toEqual({ a: '4', b: '' })
      expect(next.fieldErrors).toEqual({ b: 'Enter a number.' })
    })

    it('clears a request error', () => {
      const next = calculatorReducer(stateWith({ status: failed }), { type: 'inputChanged', field: 'b', value: '2' })

      expect(next.status).toEqual({ kind: 'idle' })
    })

    it('keeps the last result visible', () => {
      const next = calculatorReducer(stateWith({ status: success }), { type: 'inputChanged', field: 'a', value: '9' })

      expect(next.status).toBe(success)
    })

    it('leaves a request in flight alone', () => {
      const next = calculatorReducer(stateWith({ status: loading }), { type: 'inputChanged', field: 'a', value: '9' })

      expect(next.status).toBe(loading)
    })
  })

  describe('operationSelected', () => {
    it('sets the operation and clears field and request errors', () => {
      const state = stateWith({ fieldErrors: { a: 'Enter a number.' }, status: failed })

      const next = calculatorReducer(state, { type: 'operationSelected', operation: 'sqrt' })

      expect(next.operation).toBe('sqrt')
      expect(next.fieldErrors).toEqual({})
      expect(next.status).toEqual({ kind: 'idle' })
    })

    it('keeps the inputs and the last result', () => {
      const state = stateWith({ input: { a: '2', b: '3' }, status: success })

      const next = calculatorReducer(state, { type: 'operationSelected', operation: 'multiply' })

      expect(next.input).toEqual({ a: '2', b: '3' })
      expect(next.status).toBe(success)
    })
  })

  describe('submitted', () => {
    it('parses the inputs and starts a request', () => {
      const state = stateWith({ operation: 'multiply', input: { a: ' 1e3 ', b: '-2.5' } })

      const next = calculatorReducer(state, { type: 'submitted' })

      expect(next.fieldErrors).toEqual({})
      expect(next.status).toEqual({ kind: 'loading', request: { operation: 'multiply', operands: [1000, -2.5] } })
    })

    it('sends one operand for sqrt and ignores the second input', () => {
      const state = stateWith({ operation: 'sqrt', input: { a: '16', b: 'not a number' } })

      const next = calculatorReducer(state, { type: 'submitted' })

      expect(next.status).toEqual({ kind: 'loading', request: { operation: 'sqrt', operands: [16] } })
    })

    it('reports every invalid input and starts no request', () => {
      const state = stateWith({ input: { a: '', b: '1,5' } })

      const next = calculatorReducer(state, { type: 'submitted' })

      expect(next.fieldErrors).toEqual({ a: 'Enter a number.', b: 'Use a dot for decimals, like 1.5.' })
      expect(next.status).toEqual({ kind: 'idle' })
    })

    it('reports only the invalid input', () => {
      const next = calculatorReducer(stateWith({ input: { a: '2', b: 'x' } }), { type: 'submitted' })

      expect(next.fieldErrors).toEqual({ b: 'Enter a valid number, like -3.5 or 1e3.' })
    })

    it('clears a request error but keeps the last result when validation fails', () => {
      const invalid = { input: { a: '', b: '' } }

      expect(calculatorReducer(stateWith({ ...invalid, status: failed }), { type: 'submitted' }).status).toEqual({
        kind: 'idle',
      })
      expect(calculatorReducer(stateWith({ ...invalid, status: success }), { type: 'submitted' }).status).toBe(success)
    })

    it('clears old field errors when the inputs are valid', () => {
      const state = stateWith({ input: { a: '1', b: '2' }, fieldErrors: { a: 'Enter a number.' } })

      expect(calculatorReducer(state, { type: 'submitted' }).fieldErrors).toEqual({})
    })

    it('is ignored while a request is in flight', () => {
      const state = stateWith({ input: { a: '7', b: '8' }, status: loading })

      expect(calculatorReducer(state, { type: 'submitted' })).toBe(state)
    })
  })

  describe('responses', () => {
    it('stores a successful result with the request that produced it', () => {
      const next = calculatorReducer(stateWith({ status: loading }), { type: 'requestSucceeded', result: 5 })

      expect(next.status).toEqual(success)
    })

    it('stores a failure message', () => {
      const next = calculatorReducer(stateWith({ status: loading }), {
        type: 'requestFailed',
        message: 'You can’t divide by zero.',
      })

      expect(next.status).toEqual(failed)
    })

    it.each([
      ['idle', { kind: 'idle' } as Status],
      ['success', success],
      ['error', failed],
    ])('are ignored when no request is in flight (%s)', (_, status) => {
      const state = stateWith({ status })

      expect(calculatorReducer(state, { type: 'requestSucceeded', result: 1 })).toBe(state)
      expect(calculatorReducer(state, { type: 'requestFailed', message: 'x' })).toBe(state)
    })
  })
})
