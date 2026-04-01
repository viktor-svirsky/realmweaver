import { rpc } from './nakama';
import { TravelResponse } from '../types/game';

export interface NearbyPlayer {
  user_id: string;
  character_id: string;
  character_name: string;
  character_class: string;
  character_level: number;
  region_x: number;
  region_y: number;
  karma: number;
  pk_count: number;
  pvp_count: number;
  flagged: boolean;
}

export interface ChatMessage {
  id: number;
  user_id: string;
  content: string;
  created_at: string;
}

export interface TradeOffer {
  trade_id: string;
  seller_id: string;
  seller_name: string;
  offer_item_name?: string;
  offer_gold: number;
  want_gold: number;
  status: string;
}

export async function getNearbyPlayers(regionX: number, regionY: number): Promise<NearbyPlayer[]> {
  const result = await rpc<{ players: NearbyPlayer[] | null }>('get_nearby_players', {
    region_x: regionX,
    region_y: regionY,
  });
  return result.players || [];
}

export async function postChat(regionX: number, regionY: number, characterName: string, message: string): Promise<void> {
  await rpc('post_chat', {
    region_x: regionX,
    region_y: regionY,
    character_name: characterName,
    message,
  });
}

export async function getChat(regionX: number, regionY: number): Promise<ChatMessage[]> {
  const result = await rpc<{ messages: ChatMessage[] | null }>('get_chat', {
    region_x: regionX,
    region_y: regionY,
  });
  return result.messages || [];
}

export async function listTrades(regionX: number, regionY: number): Promise<TradeOffer[]> {
  const result = await rpc<{ trades: TradeOffer[] | null }>('trade_list', {
    region_x: regionX,
    region_y: regionY,
  });
  return result.trades || [];
}

export async function pvpChallenge(
  attackerCharId: string,
  defenderUserId: string,
  defenderCharId: string,
): Promise<Record<string, unknown>> {
  return rpc('pvp_challenge', {
    attacker_char_id: attackerCharId,
    defender_user_id: defenderUserId,
    defender_char_id: defenderCharId,
  });
}

export async function travel(
  characterId: string,
  regionX: number,
  regionY: number,
): Promise<TravelResponse> {
  return rpc<TravelResponse>('travel', {
    character_id: characterId,
    region_x: regionX,
    region_y: regionY,
  });
}

export async function coopHelp(
  helperCharId: string,
  targetUserId: string,
  targetCharId: string,
  helpType: 'heal' | 'buff_str' | 'buff_ac',
): Promise<{ result: string }> {
  return rpc('coop_help', {
    helper_char_id: helperCharId,
    target_user_id: targetUserId,
    target_char_id: targetCharId,
    help_type: helpType,
  });
}
