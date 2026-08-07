import { useEffect, useRef } from 'react';
import { wsClient, type HubEvent } from './wsClient';

/**
 * Subscribes to the shared hub WS connection. The latest onEvent is always
 * used, so callers don't need to memoize it.
 */
export function useWsSubscribe(onEvent: (e: HubEvent) => void) {
  const handlerRef = useRef(onEvent);
  handlerRef.current = onEvent;

  useEffect(() => wsClient.subscribe((e) => handlerRef.current(e)), []);
}
