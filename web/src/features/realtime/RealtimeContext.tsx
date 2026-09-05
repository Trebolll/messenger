import { createContext, useContext, useEffect, useRef, useCallback, type ReactNode } from 'react';
import { WS_BASE, TOKEN_KEY } from '../../shared/config';

type Handler = (event: Record<string, unknown>) => void;

type RealtimeApi = {
  subscribe: (channels: unknown[]) => void;
  on: (type: string, handler: Handler) => () => void;
  connected: boolean;
};

const RealtimeContext = createContext<RealtimeApi | null>(null);

export function RealtimeProvider({ children, enabled }: { children: ReactNode; enabled: boolean }) {
  const wsRef = useRef<WebSocket | null>(null);
  const handlers = useRef<Map<string, Set<Handler>>>(new Map());
  const connectedRef = useRef(false);

  const emit = useCallback((event: Record<string, unknown>) => {
    const type = String(event.type ?? '');
    handlers.current.get(type)?.forEach((h) => h(event));
    handlers.current.get('*')?.forEach((h) => h(event));
  }, []);

  useEffect(() => {
    if (!enabled) return;
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) return;

    const ws = new WebSocket(`${WS_BASE}/api/ws?token=${encodeURIComponent(token)}`);
    wsRef.current = ws;

    ws.onopen = () => {
      connectedRef.current = true;
    };
    ws.onmessage = (msg) => {
      try {
        emit(JSON.parse(msg.data));
      } catch {
        /* ignore */
      }
    };
    ws.onclose = () => {
      connectedRef.current = false;
    };

    const ping = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 25000);

    return () => {
      clearInterval(ping);
      ws.close();
      wsRef.current = null;
    };
  }, [enabled, emit]);

  const subscribe = useCallback((channels: unknown[]) => {
    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'subscribe', channels }));
    } else {
      const trySub = setInterval(() => {
        const w = wsRef.current;
        if (w && w.readyState === WebSocket.OPEN) {
          w.send(JSON.stringify({ type: 'subscribe', channels }));
          clearInterval(trySub);
        }
      }, 200);
      setTimeout(() => clearInterval(trySub), 5000);
    }
  }, []);

  const on = useCallback((type: string, handler: Handler) => {
    if (!handlers.current.has(type)) handlers.current.set(type, new Set());
    handlers.current.get(type)!.add(handler);
    return () => handlers.current.get(type)?.delete(handler);
  }, []);

  return (
    <RealtimeContext.Provider value={{ subscribe, on, connected: connectedRef.current }}>
      {children}
    </RealtimeContext.Provider>
  );
}

export function useRealtime() {
  const ctx = useContext(RealtimeContext);
  if (!ctx) throw new Error('RealtimeProvider missing');
  return ctx;
}
