import { create } from 'zustand';
import { Character, GamePhase, CombatState, QuickAction, NarrativeEntry } from '../types/game';

interface GameStore {
  character: Character | null;
  setCharacter: (character: Character) => void;

  phase: GamePhase;
  setPhase: (phase: GamePhase) => void;

  combat: CombatState | null;
  setCombat: (combat: CombatState | null) => void;

  narrativeEntries: NarrativeEntry[];
  addNarrativeEntry: (entry: NarrativeEntry) => void;
  clearNarrative: () => void;

  isStreaming: boolean;
  streamBuffer: string;
  setStreaming: (streaming: boolean) => void;
  appendStreamToken: (token: string) => void;
  finalizeStream: () => void;

  quickActions: QuickAction[];
  setQuickActions: (actions: QuickAction[]) => void;

  isLoading: boolean;
  setLoading: (loading: boolean) => void;

  language: string;
  setLanguage: (lang: string) => void;
}

export const useGameStore = create<GameStore>((set, get) => ({
  character: null,
  setCharacter: (character) => set({ character }),

  phase: 'exploring',
  setPhase: (phase) => set({ phase }),

  combat: null,
  setCombat: (combat) => set({ combat }),

  narrativeEntries: [],
  addNarrativeEntry: (entry) =>
    set((state) => {
      const next = [...state.narrativeEntries, entry];
      return { narrativeEntries: next.length > 200 ? next.slice(-200) : next };
    }),
  clearNarrative: () => set({ narrativeEntries: [] }),

  isStreaming: false,
  streamBuffer: '',
  setStreaming: (streaming) => set({ isStreaming: streaming }),
  appendStreamToken: (token) =>
    set((state) => ({ streamBuffer: state.streamBuffer + token })),
  finalizeStream: () => {
    const buffer = get().streamBuffer;
    if (buffer) {
      set((state) => ({
        streamBuffer: '',
        narrativeEntries: [
          ...state.narrativeEntries,
          {
            id: `stream_${Date.now()}`,
            type: 'narrative' as const,
            text: buffer,
            timestamp: Date.now(),
          },
        ],
      }));
    }
  },

  quickActions: [],
  setQuickActions: (actions) => set({ quickActions: actions }),

  isLoading: false,
  setLoading: (loading) => set({ isLoading: loading }),

  language: 'en',
  setLanguage: (lang) => set({ language: lang }),
}));
