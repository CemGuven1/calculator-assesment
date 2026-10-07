import { useEffect, useMemo, useReducer } from 'react'
import { calculate } from '../api/client.ts'
import type { Operation } from '../api/types.ts'
import { calculatorReducer, initialState, type Field } from './reducer.ts'

/**
 * Calculator state plus the actions the UI can take. The reducer decides when
 * a request starts; this hook only sends it and reports the outcome.
 */
export function useCalculator() {
  const [state, dispatch] = useReducer(calculatorReducer, initialState)
  const { status } = state

  // Send the request whenever the reducer enters the loading state. Aborting
  // on cleanup means a response that arrives after unmount is ignored.
  useEffect(() => {
    if (status.kind !== 'loading') {
      return
    }
    const controller = new AbortController()
    const { operation, operands } = status.request
    void calculate(operation, operands, controller.signal).then((outcome) => {
      if (controller.signal.aborted) {
        return
      }
      dispatch(
        outcome.ok
          ? { type: 'requestSucceeded', result: outcome.result }
          : { type: 'requestFailed', message: outcome.error.message },
      )
    })
    return () => controller.abort()
  }, [status])

  const actions = useMemo(
    () => ({
      selectOperation: (operation: Operation) => dispatch({ type: 'operationSelected', operation }),
      changeInput: (field: Field, value: string) => dispatch({ type: 'inputChanged', field, value }),
      submit: () => dispatch({ type: 'submitted' }),
    }),
    [],
  )

  return { state, ...actions }
}
