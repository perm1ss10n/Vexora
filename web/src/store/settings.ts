import { create } from 'zustand';

const SETTINGS_KEY = 'konyx-settings';

interface UiSettings {
  denseMode: boolean;
  glowEffects: boolean;
  telemetryIntervalMs: number;
  minPublishMs: number;
}

interface SettingsState extends UiSettings {
  update: (next: Partial<UiSettings>) => void;
}

const defaultSettings: UiSettings = {
  denseMode: false,
  glowEffects: true,
  telemetryIntervalMs: 5000,
  minPublishMs: 15000,
};

const loadSettings = (): UiSettings => {
  const raw = localStorage.getItem(SETTINGS_KEY);
  if (!raw) return defaultSettings;
  try {
    const parsed = JSON.parse(raw) as Partial<UiSettings>;
    return { ...defaultSettings, ...parsed };
  } catch {
    return defaultSettings;
  }
};

export const useSettingsStore = create<SettingsState>((set) => ({
  ...loadSettings(),
  update: (next) =>
    set((state) => {
      const updated = { ...state, ...next };
      const { update: _, ...persisted } = updated as SettingsState;
      localStorage.setItem(SETTINGS_KEY, JSON.stringify(persisted));
      return updated;
    }),
}));
