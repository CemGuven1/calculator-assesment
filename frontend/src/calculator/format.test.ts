import { describe, expect, it } from 'vitest'
import type { Operation } from '../api/types.ts'
import { describeCalculation, formatNumber } from './format.ts'

describe('formatNumber', () => {
  it.each([
    [5, '5'],
    [-2.5, '-2.5'],
    [0.30000000000000004, '0.3'],
    [1 / 3, '0.333333333333333'],
    [-0, '0'],
    [1e21, '1e+21'],
    [1.7976931348623157e308, '1.7976931348623157e+308'],
    [-1.7976931348623157e308, '-1.7976931348623157e+308'],
    [1.79769313486231e308, '1.79769313486231e+308'],
    [1e-7, '1e-7'],
    [123456789012345680, '123456789012346000'],
  ])('formats %d as %s', (value, text) => {
    expect(formatNumber(value)).toBe(text)
  })
})

describe('describeCalculation', () => {
  it.each<[Operation, number[], number, string, string]>([
    ['add', [2, 3], 5, '2 + 3', '5'],
    ['subtract', [5, -3], 8, '5 − (-3)', '8'],
    ['multiply', [4, 2.5], 10, '4 × 2.5', '10'],
    ['divide', [10, 4], 2.5, '10 ÷ 4', '2.5'],
    ['power', [-2, 3], -8, '(-2) ^ 3', '-8'],
    ['sqrt', [16], 4, '√16', '4'],
    ['percentage', [15, 200], 30, '15% of 200', '30'],
    ['add', [0.1, 0.2], 0.30000000000000004, '0.1 + 0.2', '0.3'],
  ])('%s %j', (operation, operands, result, expression, resultText) => {
    expect(describeCalculation({ operation, operands, result })).toEqual({ expression, result: resultText })
  })
})
