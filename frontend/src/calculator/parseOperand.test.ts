import { describe, expect, it } from 'vitest'
import { parseOperand } from './parseOperand.ts'

describe('parseOperand', () => {
  it.each([
    ['42', 42],
    ['-3.5', -3.5],
    ['+7', 7],
    ['0.1', 0.1],
    ['.5', 0.5],
    ['5.', 5],
    ['1e3', 1000],
    ['2.5E-3', 0.0025],
    ['-1e+2', -100],
    ['  12  ', 12],
    ['007', 7],
    ['0', 0],
    ['-0.000', -0],
    ['0e-500', 0],
    ['1e-310', 1e-310],
  ])('parses %j as %d', (text, value) => {
    expect(parseOperand(text)).toEqual({ ok: true, value })
  })

  it.each([
    ['', 'Enter a number.'],
    ['   ', 'Enter a number.'],
    ['1,5', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['-0,25', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['1,000', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['1,000,000', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['1 000', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['1,000.5', 'Use a dot for decimals and no thousands separators, like 1500 or 1.5.'],
    ['1..2', 'A number can only have one decimal point.'],
    ['1.2.3', 'A number can only have one decimal point.'],
    ['..', 'A number can only have one decimal point.'],
    ['abc', 'Enter a valid number, like -3.5 or 1e3.'],
    ['.', 'Enter a valid number, like -3.5 or 1e3.'],
    ['-', 'Enter a valid number, like -3.5 or 1e3.'],
    ['1e', 'Enter a valid number, like -3.5 or 1e3.'],
    ['0x10', 'Enter a valid number, like -3.5 or 1e3.'],
    ['Infinity', 'Enter a valid number, like -3.5 or 1e3.'],
    ['NaN', 'Enter a valid number, like -3.5 or 1e3.'],
    ['1e400', 'This number is too large.'],
    ['-1e400', 'This number is too large.'],
    ['1e-400', 'This number is too small.'],
    ['-0.0001e-400', 'This number is too small.'],
  ])('rejects %j', (text, error) => {
    expect(parseOperand(text)).toEqual({ ok: false, error })
  })
})
