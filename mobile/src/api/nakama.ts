import { Client, Session } from '@heroiclabs/nakama-js';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Configurable via environment or defaults for local dev
const NAKAMA_HOST = process.env.EXPO_PUBLIC_NAKAMA_HOST || 'localhost';
const NAKAMA_PORT = process.env.EXPO_PUBLIC_NAKAMA_PORT || '7350';
const NAKAMA_KEY = process.env.EXPO_PUBLIC_NAKAMA_KEY || 'defaultkey';
const NAKAMA_SSL = process.env.EXPO_PUBLIC_NAKAMA_SSL === 'true';

const DEVICE_ID_KEY = 'realmweaver_device_id';
const SESSION_KEY = 'realmweaver_session';

let client: Client;
let session: Session | null = null;

export function getClient(): Client {
  if (!client) {
    client = new Client(NAKAMA_KEY, NAKAMA_HOST, NAKAMA_PORT, NAKAMA_SSL);
  }
  return client;
}

async function getOrCreateDeviceId(): Promise<string> {
  try {
    const stored = await AsyncStorage.getItem(DEVICE_ID_KEY);
    if (stored) return stored;
  } catch {
    // AsyncStorage not available (web fallback)
  }

  // Try localStorage for web
  if (typeof window !== 'undefined' && window.localStorage) {
    const stored = window.localStorage.getItem(DEVICE_ID_KEY);
    if (stored) return stored;
  }

  const id = `device_${Date.now()}_${Math.random().toString(36).slice(2)}`;

  try {
    await AsyncStorage.setItem(DEVICE_ID_KEY, id);
  } catch {
    // Fallback to localStorage for web
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.setItem(DEVICE_ID_KEY, id);
    }
  }

  return id;
}

export async function authenticate(): Promise<Session> {
  const c = getClient();

  // Try to restore session
  try {
    let stored: string | null = null;
    try {
      stored = await AsyncStorage.getItem(SESSION_KEY);
    } catch {
      if (typeof window !== 'undefined' && window.localStorage) {
        stored = window.localStorage.getItem(SESSION_KEY);
      }
    }

    if (stored) {
      const parsed = JSON.parse(stored);
      const now = Math.floor(Date.now() / 1000);
      // Check if token is still valid (with 60s buffer)
      if (parsed.expires_at && parsed.expires_at > now + 60) {
        // Reconstruct session — nakama-js Session.restore
        session = Session.restore(parsed.token, parsed.refresh_token);
        if (session && !session.isexpired(now)) {
          return session;
        }
      }
      // Token expired but refresh_token exists — try to refresh
      if (parsed.refresh_token) {
        try {
          const restored = Session.restore(parsed.token, parsed.refresh_token);
          session = await c.sessionRefresh(restored);
          // Persist refreshed session
          const refreshedData = JSON.stringify({
            token: session.token,
            refresh_token: session.refresh_token,
            expires_at: session.expires_at,
          });
          try {
            await AsyncStorage.setItem(SESSION_KEY, refreshedData);
          } catch {
            if (typeof window !== 'undefined' && window.localStorage) {
              window.localStorage.setItem(SESSION_KEY, refreshedData);
            }
          }
          return session;
        } catch {
          // Refresh failed, fall through to fresh auth
        }
      }
    }
  } catch {
    // Failed to restore, will re-authenticate
  }

  // Fresh authentication
  const deviceId = await getOrCreateDeviceId();
  session = await c.authenticateDevice(deviceId, true);

  // Persist session
  const sessionData = JSON.stringify({
    token: session.token,
    refresh_token: session.refresh_token,
    expires_at: session.expires_at,
  });
  try {
    await AsyncStorage.setItem(SESSION_KEY, sessionData);
  } catch {
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.setItem(SESSION_KEY, sessionData);
    }
  }

  return session;
}

export function getSession(): Session | null {
  return session;
}

export async function rpc<T>(name: string, payload: Record<string, unknown>): Promise<T> {
  const s = getSession();
  if (!s) throw new Error('Not authenticated');

  const baseUrl = `http://${NAKAMA_HOST}:${NAKAMA_PORT}`;
  const resp = await fetch(`${baseUrl}/v2/rpc/${name}?unwrap`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${s.token}`,
    },
    body: JSON.stringify(payload),
  });

  if (!resp.ok) {
    const errorText = await resp.text();
    throw new Error(`RPC ${name} failed: ${resp.status} ${errorText}`);
  }

  const text = await resp.text();
  return JSON.parse(text) as T;
}
