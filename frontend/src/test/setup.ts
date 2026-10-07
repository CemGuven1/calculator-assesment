import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Testing Library only auto-unmounts when Vitest globals are enabled; we import
// test functions explicitly instead, so unmount rendered trees here.
afterEach(() => {
  cleanup()
})
