import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ResultPanel } from './ResultPanel.tsx'

describe('ResultPanel', () => {
  // Checked in a browser at 360px: these sizes keep each number on one line.
  it.each([
    [5, '5', 'result-value'],
    [Math.SQRT2, '1.4142135623731', 'result-value result-value-long'],
    [-1.7976931348623157e308, '-1.7976931348623157e+308', 'result-value result-value-xlong'],
  ])('shrinks long results to fit (%d)', (result, text, className) => {
    render(
      <ResultPanel status={{ kind: 'success', calculation: { operation: 'multiply', operands: [result, 1], result } }} />,
    )

    expect(screen.getByText(text)).toHaveAttribute('class', className)
  })
})
