import { describe, it, expect, vi } from 'vitest'
import { renderWithProviders } from '../test/renderWithProviders'
import { screen } from '@testing-library/react'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (k: string) => k, i18n: { language: 'en' } }),
  initReactI18next: { type: '3rdParty', init: () => {} },
}))
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => vi.fn() }
})
vi.mock('../features/auth/AuthProvider', () => ({
  useAuth: () => ({ login: vi.fn(), logout: vi.fn(), isLoading: false, isAuthenticated: false, currentUser: null, hasPermission: () => false, requirePermission: () => false, refreshToken: vi.fn() }),
  AuthProvider: ({ children }: any) => children,
}))

import Login from '../pages/Login'

describe('Login Page', () => {
  it('renders email input', () => {
    renderWithProviders(<Login />)
    expect(screen.getByPlaceholderText('login.emailPlaceholder')).toBeDefined()
  })
  it('renders password input', () => {
    renderWithProviders(<Login />)
    expect(screen.getByPlaceholderText('login.passwordPlaceholder')).toBeDefined()
  })
  it('renders admin title', () => {
    renderWithProviders(<Login />)
    expect(screen.getByText('login.title')).toBeDefined()
  })
})
