import axios from "axios";
import AsyncStorage from "@react-native-async-storage/async-storage";

const API_URL = process.env.EXPO_PUBLIC_API_URL || "http://10.0.2.2:80";

export const api = axios.create({
  baseURL: API_URL,
  headers: { "Content-Type": "application/json" },
  timeout: 30000,
});

api.interceptors.request.use(async (config) => {
  const token = await AsyncStorage.getItem("zu_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Sincronización offline
const OFFLINE_QUEUE_KEY = "zu_offline_queue";

export async function queueOfflineAction(action: { method: string; url: string; data: unknown }) {
  const queue = JSON.parse((await AsyncStorage.getItem(OFFLINE_QUEUE_KEY)) || "[]");
  queue.push({ ...action, queued_at: new Date().toISOString() });
  await AsyncStorage.setItem(OFFLINE_QUEUE_KEY, JSON.stringify(queue));
}

export async function syncOfflineQueue() {
  const queue = JSON.parse((await AsyncStorage.getItem(OFFLINE_QUEUE_KEY)) || "[]");
  if (queue.length === 0) return { synced: 0 };

  let synced = 0;
  const remaining: any[] = [];
  for (const action of queue) {
    try {
      await api.request({ method: action.method, url: action.url, data: action.data });
      synced++;
    } catch {
      remaining.push(action);
    }
  }
  await AsyncStorage.setItem(OFFLINE_QUEUE_KEY, JSON.stringify(remaining));
  return { synced, remaining: remaining.length };
}

export default api;
