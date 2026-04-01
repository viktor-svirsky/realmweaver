import { Socket, MatchData } from '@heroiclabs/nakama-js';
import { getClient, getSession, rpc } from './nakama';
import { OpCode, GameState, QuickAction, ActionResult, NarrativeEntry } from '../types/game';
import { useGameStore } from '../state/gameStore';

let socket: Socket | null = null;
let currentMatchId: string | null = null;
let idCounter = 0;
function nextId(prefix: string): string {
  return `${prefix}_${Date.now()}_${++idCounter}`;
}

export async function connectSocket(): Promise<Socket> {
  // Clean up previous socket if any
  if (socket) {
    try {
      socket.ondisconnect = () => {};
      socket.onmatchdata = () => {};
      socket.disconnect(false);
    } catch {
      // ignore cleanup errors
    }
    socket = null;
    currentMatchId = null;
  }

  const client = getClient();
  const session = getSession();
  if (!session) throw new Error('Not authenticated');

  socket = client.createSocket(false, false);
  await socket.connect(session, false);

  socket.onmatchdata = handleMatchData;
  socket.ondisconnect = () => {
    console.log('Socket disconnected');
  };

  return socket;
}

export async function createMatch(characterId: string, language?: string): Promise<string> {
  if (!socket) throw new Error('Socket not connected');

  const lang = language || useGameStore.getState().language || 'en';
  console.log('Creating match with language:', lang);

  // Create match server-side via RPC
  const result = await rpc<{ match_id: string }>('start_game', {
    character_id: characterId,
    language: lang,
  });

  // Join the match via WebSocket
  const match = await socket.joinMatch(result.match_id);
  currentMatchId = match.match_id;
  console.log('Joined match:', currentMatchId);
  return currentMatchId;
}

export function sendAction(action: string, options?: { target?: string; itemId?: string; text?: string }): void {
  if (!socket || !currentMatchId) {
    console.log('Cannot send action: no socket or match', { socket: !!socket, matchId: currentMatchId });
    return;
  }

  const payload = {
    action,
    target: options?.target || '',
    item_id: options?.itemId || '',
    text: options?.text || '',
  };

  console.log('Sending action:', payload);
  socket.sendMatchState(currentMatchId, OpCode.PLAYER_ACTION, JSON.stringify(payload));
}

function handleMatchData(matchData: MatchData): void {
  const store = useGameStore.getState();

  let data: Record<string, unknown>;
  try {
    const decoded = typeof matchData.data === 'string'
      ? matchData.data
      : new TextDecoder().decode(matchData.data as Uint8Array);
    data = JSON.parse(decoded);
  } catch {
    console.error('Failed to parse match data');
    return;
  }

  console.log('Match data received, opcode:', matchData.op_code);

  switch (Number(matchData.op_code)) {
    case OpCode.GAME_STATE: {
      const state = data as unknown as GameState;
      if (state.character) store.setCharacter(state.character);
      store.setPhase(state.phase);
      if (state.combat) store.setCombat(state.combat);
      break;
    }

    case OpCode.NARRATIVE: {
      const entry: NarrativeEntry = {
        id: nextId('narr'),
        type: 'narrative',
        text: (data as { text: string }).text,
        timestamp: Date.now(),
      };
      store.addNarrativeEntry(entry);
      store.setStreaming(false);
      store.setLoading(false);
      break;
    }

    case OpCode.MECHANICAL: {
      const result = data as unknown as ActionResult;
      const entry: NarrativeEntry = {
        id: nextId('mech'),
        type: 'mechanical',
        text: result.details,
        timestamp: Date.now(),
      };
      store.addNarrativeEntry(entry);

      if (result.victory) {
        store.setCombat(null);
        if (result.xp_gained) {
          store.addNarrativeEntry({
            id: nextId('xp'),
            type: 'system',
            text: `+${result.xp_gained} XP${result.leveled_up ? ' LEVEL UP!' : ''}`,
            timestamp: Date.now(),
          });
        }
      }
      break;
    }

    case OpCode.STREAM_TOKEN:
    case OpCode.STREAM_END:
      // No longer used — server sends complete narrative
      break;

    case OpCode.QUICK_ACTIONS: {
      const actions = data as unknown as QuickAction[];
      store.setQuickActions(actions);
      break;
    }

    case OpCode.ERROR: {
      console.error('Server error:', (data as { error: string }).error);
      store.addNarrativeEntry({
        id: nextId('err'),
        type: 'system',
        text: `Error: ${(data as { error: string }).error}`,
        timestamp: Date.now(),
      });
      break;
    }
  }
}

export function getSocket(): Socket | null {
  return socket;
}

export function getCurrentMatchId(): string | null {
  return currentMatchId;
}
