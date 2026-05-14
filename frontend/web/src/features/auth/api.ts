import { createClient } from '@connectrpc/connect';
import { transport } from '../_shared/transport';
import { AuthService } from '@antclaw/proto/antclaw/v1/auth_pb';
import { create } from '@bufbuild/protobuf';
import { LoginRequestSchema, RegisterRequestSchema } from '@antclaw/proto/antclaw/v1/auth_pb';
import { getDeviceId } from '../_shared/deviceInfo';

const authClient = createClient(AuthService, transport);

export async function login(email: string, password: string) {
  const req = create(LoginRequestSchema, {
    email,
    password,
    client: {
      userAgent: navigator.userAgent,
      ipAddress: '',
      deviceId: getDeviceId(),
    },
  });
  const resp = await authClient.login(req);
  localStorage.setItem('token', resp.accessToken);
  localStorage.setItem('refreshToken', resp.refreshToken);
  return resp;
}

export async function register(email: string, password: string, displayName?: string) {
  const req = create(RegisterRequestSchema, {
    email,
    password,
    displayName: displayName || email.split('@')[0],
    locale: 0, // LOCALE_UNSPECIFIED
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    client: {
      userAgent: navigator.userAgent,
      ipAddress: '',
      deviceId: getDeviceId(),
    },
    idempotencyKey: crypto.randomUUID(),
  });
  const resp = await authClient.register(req);
  localStorage.setItem('token', resp.accessToken);
  localStorage.setItem('refreshToken', resp.refreshToken);
  return resp;
}

export function getToken(): string | null {
  return localStorage.getItem('token');
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}

export function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('refreshToken');
}
