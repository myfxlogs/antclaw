import { ReactNode } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'

function Wrapper({ children }: { children: ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>
}

export function renderWithProviders(ui: ReactNode, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: Wrapper, ...options })
}

export { Wrapper as TestProviders }
