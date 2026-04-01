export type Class = 'warrior' | 'mage' | 'rogue' | 'cleric' | 'ranger' | 'paladin' | 'necromancer' | 'berserker';

export type GamePhase = 'exploring' | 'in_combat' | 'in_dialogue' | 'in_shop' | 'traveling';

export type ItemType = 'weapon' | 'armor' | 'consumable' | 'misc';
export type EquipSlot = 'weapon' | 'offhand' | 'armor' | 'helmet' | 'boots' | 'ring1' | 'ring2' | 'amulet';

export interface Stats {
  str: number;
  dex: number;
  con: number;
  int: number;
  wis: number;
  cha: number;
}

export interface Item {
  id: string;
  name: string;
  description: string;
  type: ItemType;
  slot?: EquipSlot;
  damage_dice?: number;
  damage_count?: number;
  ac_bonus?: number;
  heal_amount?: number;
  weight: number;
  value: number;
}

export interface Equipment {
  weapon?: Item;
  offhand?: Item;
  armor?: Item;
  helmet?: Item;
  boots?: Item;
  ring1?: Item;
  ring2?: Item;
  amulet?: Item;
}

export interface Character {
  id: string;
  name: string;
  class: Class;
  level: number;
  xp: number;
  stats: Stats;
  hp: number;
  max_hp: number;
  mana: number;
  max_mana: number;
  ac: number;
  gold: number;
  karma: number;
  pk_count: number;
  pvp_count: number;
  flagged: boolean;
  flagged_until: number;
  equipment: Equipment;
  inventory: Item[];
  region_x: number;
  region_y: number;
}

export type PKTitle = 'Innocent' | 'Outlaw' | 'Wanted' | 'Serial Killer';

export function getNameColor(karma: number, flagged: boolean): string {
  if (flagged) return '#9b59b6'; // purple — attacked someone
  if (karma > 0) return '#c0392b'; // red — PK karma
  return '#ffffff'; // white — clean
}

export function getPKTitle(karma: number, pkCount: number): PKTitle {
  if (karma >= 5000) return 'Serial Killer';
  if (karma >= 1000) return 'Wanted';
  if (karma > 0) return 'Outlaw';
  return 'Innocent';
}

export interface Enemy {
  id: string;
  name: string;
  hp: number;
  max_hp: number;
  ac: number;
}

export interface CombatState {
  enemies: Enemy[];
  turn_order: string[];
  current_turn: number;
  round: number;
}

export interface QuickAction {
  id: string;
  label: string;
  icon?: string;
}

export interface ActionResult {
  action: string;
  actor: string;
  target?: string;
  roll?: number;
  modifier?: number;
  dc?: number;
  hit?: boolean;
  damage?: number;
  success: boolean;
  details: string;
  hp_remaining?: number;
  enemy_defeated?: boolean;
  combat_ended?: boolean;
  victory?: boolean;
  xp_gained?: number;
  leveled_up?: boolean;
  loot_dropped?: Item[];
}

export interface NarrativeEntry {
  id: string;
  type: 'narrative' | 'mechanical' | 'system';
  text: string;
  timestamp: number;
}

export interface GameState {
  character: Character | null;
  phase: GamePhase;
  combat: CombatState | null;
}

// Travel system types

export interface RegionData {
  id: number;
  x: number;
  y: number;
  biome: string;
  difficulty: number;
  name: string;
  description: string;
  lore: string;
  structures: Structure[];
}

export interface Structure {
  name: string;
  type: string;
  description: string;
}

export interface RegionNPC {
  id: number;
  region_id: number;
  name: string;
  race: string;
  occupation: string;
  location_detail: string;
}

export interface TravelResponse {
  region: RegionData;
  travel_time: number;
  narrative: string;
  npcs: RegionNPC[];
}

// Travel times by biome (seconds)
export const BIOME_TRAVEL_TIMES: Record<string, number> = {
  plains: 3,
  farmlands: 3,
  forest: 5,
  hills: 7,
  foothills: 7,
  mountains: 10,
  swamp: 8,
  marsh: 8,
  desert: 8,
  wastes: 8,
  snow: 9,
  ice: 9,
  coast: 5,
};

export function getTravelTime(biome: string): number {
  return BIOME_TRAVEL_TIMES[biome] ?? 6;
}

export function getBiomeIcon(biome: string): string {
  const icons: Record<string, string> = {
    plains: '\u{1F33E}',
    farmlands: '\u{1F33E}',
    forest: '\u{1F332}',
    hills: '\u{26F0}\uFE0F',
    foothills: '\u{26F0}\uFE0F',
    mountains: '\u{1F3D4}\uFE0F',
    swamp: '\u{1F40A}',
    marsh: '\u{1F40A}',
    desert: '\u{1F3DC}\uFE0F',
    wastes: '\u{1F3DC}\uFE0F',
    snow: '\u{2744}\uFE0F',
    ice: '\u{2744}\uFE0F',
    coast: '\u{1F30A}',
  };
  return icons[biome] ?? '?';
}

// OpCodes matching the server
export const OpCode = {
  PLAYER_ACTION: 1,
  NARRATIVE: 2,
  GAME_STATE: 3,
  MECHANICAL: 4,
  STREAM_TOKEN: 5,
  STREAM_END: 6,
  ERROR: 7,
  QUICK_ACTIONS: 8,
} as const;
