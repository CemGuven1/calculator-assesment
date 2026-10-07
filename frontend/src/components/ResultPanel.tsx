import { describeCalculation } from '../calculator/format.ts'
import type { Status } from '../calculator/reducer.ts'

/**
 * Shows the latest request's status. The status region is always rendered so
 * screen readers announce its changes; errors use role="alert" instead, which
 * is announced as soon as it appears.
 */
export function ResultPanel({ status }: { status: Status }) {
  return (
    <section className="result-panel" aria-labelledby="result-heading">
      <h2 id="result-heading" className="result-heading">
        Result
      </h2>
      <div role="status">
        {status.kind === 'loading' && <p className="result-hint">Calculating…</p>}
        {status.kind === 'success' && <Result {...describeCalculation(status.calculation)} />}
      </div>
      {status.kind === 'error' && (
        <p role="alert" className="result-error">
          {status.message}
        </p>
      )}
    </section>
  )
}

function Result({ expression, result }: { expression: string; result: string }) {
  // Long numbers get a smaller font so they stay on one line, even at 360px.
  let size = ''
  if (result.length > 18) {
    size = ' result-value-xlong'
  } else if (result.length > 11) {
    size = ' result-value-long'
  }
  return (
    <>
      <p className="result-expression">{expression} =</p>
      {/* Keeps the region's text "2 + 3 = 5" rather than "2 + 3 =5". */}{' '}
      <p className={`result-value${size}`}>{result}</p>
    </>
  )
}
