/**
 * 极简类型化事件总线（mock 桥接与测试用）。
 * 事件名 → payload 一一对应；on 返回取消订阅函数。
 */
export type Unsubscribe = () => void;

export interface Emitter<Events extends object> {
  emit<K extends keyof Events & string>(name: K, payload: Events[K]): void;
  on<K extends keyof Events & string>(name: K, cb: (payload: Events[K]) => void): Unsubscribe;
}

export function createEmitter<Events extends object>(): Emitter<Events> {
  const listeners = new Map<string, Set<(payload: unknown) => void>>();

  return {
    emit(name, payload) {
      for (const cb of listeners.get(name) ?? []) cb(payload);
    },
    on(name, cb) {
      const set = listeners.get(name) ?? new Set();
      set.add(cb as (payload: unknown) => void);
      listeners.set(name, set);
      return () => {
        set.delete(cb as (payload: unknown) => void);
      };
    },
  };
}
