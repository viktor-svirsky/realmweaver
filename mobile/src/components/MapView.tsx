import React, { useState, useEffect, useCallback } from 'react';
import { View, Text, Pressable, StyleSheet, ScrollView, ActivityIndicator, Alert } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { sendAction } from '../api/socket';
import { getNearbyPlayers, NearbyPlayer, travel } from '../api/social';
import { getNameColor, getPKTitle, getTravelTime, getBiomeIcon, NarrativeEntry } from '../types/game';
import PlayerInteraction from './PlayerInteraction';

interface Props {
  onClose: () => void;
}

interface HexTile {
  q: number;
  r: number;
  name?: string;
  biome?: string;
  icon?: string;
  known: boolean;
  isPlayer?: boolean;
  difficulty?: number;
}

interface TownLocation {
  id: string;
  name: string;
  icon: string;
  desc: string;
  action: string;
  danger?: boolean;
}

const townLocations: TownLocation[] = [
  { id: 'tavern', name: "Wanderer's Rest", icon: '\u{1F37A}', desc: 'Marta the innkeeper', action: 'tavern' },
  { id: 'forge', name: 'Ironheart Forge', icon: '\u{2692}\uFE0F', desc: 'Theron the dwarf', action: 'forge' },
  { id: 'square', name: 'Town Square', icon: '\u{26F2}', desc: 'Pip the merchant', action: 'square' },
  { id: 'chapel', name: 'Chapel of Dawn', icon: '\u{26EA}', desc: 'Sister Lina', action: 'chapel' },
  { id: 'cave', name: 'Goblin Cave', icon: '\u{1F480}', desc: 'Danger!', action: 'enter_dungeon', danger: true },
];

const npcs = [
  { name: 'Marta', action: 'talk_marta', icon: '\u{1F469}\u200D\u{1F373}', role: 'Innkeeper' },
  { name: 'Theron', action: 'talk_theron', icon: '\u{1F9D4}', role: 'Blacksmith' },
  { name: 'Elder Corin', action: 'talk_corin', icon: '\u{1F474}', role: 'Mayor' },
  { name: 'Sister Lina', action: 'talk_lina', icon: '\u{1F64F}', role: 'Cleric' },
  { name: 'Pip', action: 'talk_pip', icon: '\u{1F9D2}', role: 'Merchant' },
];

// Generate hex grid with axial coordinates (q, r) for 4 rings (61 hexes)
function generateHexGrid(): HexTile[] {
  const tiles: HexTile[] = [];
  const radius = 4;

  // Biome hints for rings 1-3 (ring 4 is pure fog)
  const biomeHints: Record<string, { icon: string; biome: string }> = {
    // Ring 1
    '0,-1':  { icon: '\u{1F332}', biome: 'Deep Forest' },
    '1,-1':  { icon: '\u{26F0}\uFE0F', biome: 'Foothills' },
    '1,0':   { icon: '\u{1F333}', biome: 'Forest Edge' },
    '0,1':   { icon: '\u{1F33E}', biome: 'Farmlands' },
    '-1,1':  { icon: '\u{1F30A}', biome: 'River Valley' },
    '-1,0':  { icon: '\u{1F32B}\uFE0F', biome: 'Misty Woods' },
    // Ring 2
    '0,-2':  { icon: '\u{1F3D4}\uFE0F', biome: 'Mountain Pass' },
    '1,-2':  { icon: '\u{1F30B}', biome: 'Volcanic Ridge' },
    '2,-2':  { icon: '\u{1F3DC}\uFE0F', biome: 'Dry Plateau' },
    '2,-1':  { icon: '\u{1F344}', biome: 'Mushroom Grotto' },
    '2,0':   { icon: '\u{1F335}', biome: 'Cactus Wastes' },
    '1,1':   { icon: '\u{1F33B}', biome: 'Sunflower Plains' },
    '0,2':   { icon: '\u{1F3F0}', biome: 'Ruined Keep' },
    '-1,2':  { icon: '\u{1F409}', biome: 'Dragon Marsh' },
    '-2,2':  { icon: '\u{1F4A0}', biome: 'Crystal Lake' },
    '-2,1':  { icon: '\u{1F578}\uFE0F', biome: 'Spider Woods' },
    '-2,0':  { icon: '\u{2744}\uFE0F', biome: 'Frozen Glade' },
    '-1,-1': { icon: '\u{1F343}', biome: 'Whispering Pines' },
    // Ring 3
    '0,-3':  { icon: '\u{2601}\uFE0F', biome: 'Cloud Peaks' },
    '3,-3':  { icon: '\u{1F525}', biome: 'Ember Wastes' },
    '3,0':   { icon: '\u{1F3DD}\uFE0F', biome: 'Lost Oasis' },
    '0,3':   { icon: '\u{1F47B}', biome: 'Haunted Ruins' },
    '-3,3':  { icon: '\u{1F30A}', biome: 'Stormy Shore' },
    '-3,0':  { icon: '\u{2744}\uFE0F', biome: 'Ice Caverns' },
  };

  for (let q = -radius; q <= radius; q++) {
    for (let r = -radius; r <= radius; r++) {
      const s = -q - r;
      if (Math.abs(s) > radius) continue;

      const isCenter = q === 0 && r === 0;
      const ring = Math.max(Math.abs(q), Math.abs(r), Math.abs(s));

      const tile: HexTile = { q, r, known: isCenter, isPlayer: isCenter };

      if (isCenter) {
        tile.name = 'Millhaven';
        tile.biome = 'forest';
        tile.icon = '\u{1F3E1}';
        tile.difficulty = 1;
      } else {
        const key = `${q},${r}`;
        if (biomeHints[key]) {
          tile.icon = biomeHints[key].icon;
          tile.name = biomeHints[key].biome;
        }
      }

      tiles.push(tile);
    }
  }
  return tiles;
}

// Convert axial hex coords to pixel position (pointy-top hex for better vertical alignment)
const HEX_SIZE = 32;
const HEX_W = Math.sqrt(3) * HEX_SIZE;
const HEX_H = HEX_SIZE * 2;

function hexToPixel(q: number, r: number): { x: number; y: number } {
  // Pointy-top hex layout — better vertical alignment
  const x = HEX_SIZE * (Math.sqrt(3) * q + Math.sqrt(3) / 2 * r);
  const y = HEX_SIZE * (3 / 2 * r);
  return { x, y };
}

export default function MapView({ onClose }: Props) {
  const character = useGameStore((s) => s.character);
  const isLoading = useGameStore((s) => s.isLoading);
  const [selectedHex, setSelectedHex] = useState<HexTile | null>(null);
  const [nearbyPlayers, setNearbyPlayers] = useState<NearbyPlayer[]>([]);
  const [selectedPlayer, setSelectedPlayer] = useState<NearbyPlayer | null>(null);
  const [isTraveling, setIsTraveling] = useState(false);
  const [travelConfirm, setTravelConfirm] = useState<HexTile | null>(null);

  useEffect(() => {
    if (character) {
      getNearbyPlayers(character.region_x, character.region_y)
        .then(setNearbyPlayers)
        .catch(() => {});
    }
  }, [character?.region_x, character?.region_y]);

  if (!character) return null;

  const hexTiles = generateHexGrid();

  function handleLocationPress(loc: TownLocation) {
    if (isLoading) return;
    useGameStore.getState().setLoading(true);
    sendAction(loc.action);
    setTimeout(() => useGameStore.getState().setLoading(false), 15000);
    onClose();
  }

  function handleNPCPress(npc: typeof npcs[0]) {
    if (isLoading) return;
    useGameStore.getState().setLoading(true);
    sendAction(npc.action);
    setTimeout(() => useGameStore.getState().setLoading(false), 15000);
    onClose();
  }

  function handleHexTravel(tile: HexTile) {
    if (isLoading || isTraveling || tile.isPlayer) return;
    // Only allow travel to adjacent hexes (ring 1 from player position)
    const ring = Math.max(Math.abs(tile.q), Math.abs(tile.r), Math.abs(-tile.q - tile.r));
    if (ring !== 1) return;
    setTravelConfirm(tile);
  }

  const confirmTravel = useCallback(async () => {
    if (!travelConfirm) return;
    const char = useGameStore.getState().character;
    if (!char) return;
    setIsTraveling(true);
    setTravelConfirm(null);

    const store = useGameStore.getState();
    try {
      // The hex coords in the grid are relative to center (0,0).
      // Translate to world coords: char.region_x + tile.q, char.region_y + tile.r
      const destX = char.region_x + travelConfirm.q;
      const destY = char.region_y + travelConfirm.r;

      const result = await travel(char.id, destX, destY);

      // Show travel narration
      store.addNarrativeEntry({
        id: `travel_${Date.now()}`,
        type: 'narrative' as const,
        text: result.narrative,
        timestamp: Date.now(),
      });

      // Update character position in store
      store.setCharacter({
        ...char,
        region_x: destX,
        region_y: destY,
      });

      // Close map after short delay to show the result
      setTimeout(() => {
        setIsTraveling(false);
        onClose();
      }, result.travel_time * 1000);
    } catch (err) {
      setIsTraveling(false);
      store.addNarrativeEntry({
        id: `travel_err_${Date.now()}`,
        type: 'system' as const,
        text: `Travel failed: ${err instanceof Error ? err.message : 'Unknown error'}`,
        timestamp: Date.now(),
      });
    }
  }, [travelConfirm, onClose]);

  // Calculate grid bounds for centering
  const positions = hexTiles.map(t => hexToPixel(t.q, t.r));
  const minX = Math.min(...positions.map(p => p.x));
  const maxX = Math.max(...positions.map(p => p.x));
  const minY = Math.min(...positions.map(p => p.y));
  const maxY = Math.max(...positions.map(p => p.y));
  const gridW = maxX - minX + HEX_W;
  const gridH = maxY - minY + HEX_H;
  const offsetX = -minX + HEX_W * 0.5;
  const offsetY = -minY + HEX_H * 0.5;

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <View>
          <Text style={styles.title}>{'\u{1F5FA}\uFE0F'} World Map</Text>
          <Text style={styles.subtitle}>Current: Millhaven {'\u{1F333}'} Forest</Text>
        </View>
        <Pressable style={styles.closeButton} onPress={onClose}>
          <Text style={styles.closeText}>Close</Text>
        </Pressable>
      </View>

      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Hex Grid */}
        <View style={[styles.hexGrid, { width: gridW + 20, height: gridH + 20 }]}>
          {hexTiles.map((tile) => {
            const pos = hexToPixel(tile.q, tile.r);
            const ring = Math.max(Math.abs(tile.q), Math.abs(tile.r), Math.abs(-tile.q - tile.r));
            const isSelected = selectedHex?.q === tile.q && selectedHex?.r === tile.r;

            return (
              <Pressable
                key={`${tile.q},${tile.r}`}
                style={[
                  styles.hexOuter,
                  {
                    left: pos.x + offsetX - HEX_W / 2,
                    top: pos.y + offsetY - HEX_H / 2,
                    width: HEX_W,
                    height: HEX_H,
                  },
                ]}
                onPress={() => {
                  if (tile.isPlayer) {
                    setSelectedHex(isSelected ? null : tile);
                  } else if (ring >= 1 && ring <= 4) {
                    handleHexTravel(tile);
                  }
                }}
              >
                <View
                  style={[
                    styles.hex,
                    tile.isPlayer && styles.hexPlayer,
                    tile.known && !tile.isPlayer && styles.hexKnown,
                    ring <= 2 && !tile.known && !tile.isPlayer && styles.hexHinted,
                    ring === 3 && !tile.name && styles.hexDistant,
                    ring >= 4 && styles.hexFog,
                    isSelected && styles.hexSelected,
                  ]}
                >
                  <Text style={[styles.hexIcon, ring >= 3 && !tile.name && styles.hexIconFog]}>
                    {tile.icon || '?'}
                  </Text>
                  {tile.name && ring <= 2 && (
                    <Text style={[styles.hexName, ring >= 1 && styles.hexNameHinted]}>
                      {tile.name}
                    </Text>
                  )}
                  {tile.isPlayer && (
                    <Text style={styles.hexYou}>{'\u{1F9DD}'}</Text>
                  )}
                  {ring >= 1 && ring <= 3 && !tile.isPlayer && (
                    <Text style={styles.hexLevel}>Lv {ring * 2}-{ring * 3 + 1}</Text>
                  )}
                </View>
              </Pressable>
            );
          })}
        </View>

        <Text style={styles.hexHint}>
          {selectedHex?.isPlayer
            ? 'Tap a location below to explore'
            : 'Tap Millhaven to see locations'}
        </Text>

        {/* Town Locations (shown when center hex selected) */}
        {selectedHex?.isPlayer && (
          <>
            <Text style={styles.sectionTitle}>Millhaven Locations</Text>
            <View style={styles.locationGrid}>
              {townLocations.map((loc) => (
                <Pressable
                  key={loc.id}
                  style={[
                    styles.locationCard,
                    loc.danger && styles.locationDanger,
                    isLoading && styles.disabled,
                  ]}
                  onPress={() => handleLocationPress(loc)}
                  disabled={isLoading}
                >
                  <Text style={styles.locationIcon}>{loc.icon}</Text>
                  <Text style={styles.locationName}>{loc.name}</Text>
                  <Text style={[styles.locationDesc, loc.danger && styles.dangerText]}>
                    {loc.desc}
                  </Text>
                </Pressable>
              ))}
            </View>

            <Text style={styles.sectionTitle}>People</Text>
            <View style={styles.npcRow}>
              {npcs.map((npc) => (
                <Pressable
                  key={npc.name}
                  style={[styles.npcCard, isLoading && styles.disabled]}
                  onPress={() => handleNPCPress(npc)}
                  disabled={isLoading}
                >
                  <Text style={styles.npcIcon}>{npc.icon}</Text>
                  <Text style={styles.npcName}>{npc.name}</Text>
                  <Text style={styles.npcRole}>{npc.role}</Text>
                </Pressable>
              ))}
            </View>
          </>
        )}

        {/* Nearby Players */}
        {nearbyPlayers.length > 0 && (
          <>
            <Text style={styles.sectionTitle}>{'\u{1F465}'} Nearby Players ({nearbyPlayers.length})</Text>
            <Text style={styles.playerHint}>Tap a player to interact</Text>
            <View style={styles.npcRow}>
              {nearbyPlayers.map((player) => (
                <Pressable key={player.user_id} style={styles.playerCard} onPress={() => setSelectedPlayer(player)}>
                  <Text style={styles.playerIcon}>
                    {player.character_class === 'warrior' ? '\u{1F6E1}\uFE0F' :
                     player.character_class === 'mage' ? '\u{1F9D9}' :
                     player.character_class === 'rogue' ? '\u{1F977}' : '\u{1F64F}'}
                  </Text>
                  <Text style={styles.playerName}>{player.character_name}</Text>
                  <Text style={[styles.playerKarma, { color: getNameColor(player.karma, player.flagged) }]}>
                    {getPKTitle(player.karma, player.pk_count)}
                  </Text>
                  <Text style={styles.playerInfo}>Lv{player.character_level} {player.character_class}</Text>
                  {player.region_x === (character?.region_x ?? 0) && player.region_y === (character?.region_y ?? 0) && (
                    <Text style={styles.playerHere}>Here</Text>
                  )}
                </Pressable>
              ))}
            </View>
          </>
        )}

        {nearbyPlayers.length === 0 && (
          <Text style={styles.noPlayersText}>No other players nearby</Text>
        )}

        {/* Legend */}
        <View style={styles.legend}>
          <Text style={styles.legendTitle}>Legend</Text>
          <View style={styles.legendRow}>
            <View style={[styles.legendDot, { backgroundColor: '#2a4a2a' }]} />
            <Text style={styles.legendText}>Explored</Text>
            <View style={[styles.legendDot, { backgroundColor: '#1e2a3a' }]} />
            <Text style={styles.legendText}>Hinted</Text>
            <View style={[styles.legendDot, { backgroundColor: '#111' }]} />
            <Text style={styles.legendText}>Unknown</Text>
          </View>
        </View>
      </ScrollView>

      {/* Travel confirmation dialog */}
      {travelConfirm && (
        <View style={styles.travelOverlay}>
          <View style={styles.travelDialog}>
            <Text style={styles.travelTitle}>Travel to {travelConfirm.name || 'Unknown Region'}?</Text>
            <Text style={styles.travelBiome}>
              {travelConfirm.icon || '?'} {travelConfirm.name || 'Unexplored'}
            </Text>
            <Text style={styles.travelTime}>
              Estimated travel time: ~{getTravelTime(travelConfirm.biome || '')}s
            </Text>
            <View style={styles.travelButtons}>
              <Pressable style={styles.travelCancel} onPress={() => setTravelConfirm(null)}>
                <Text style={styles.travelCancelText}>Cancel</Text>
              </Pressable>
              <Pressable style={styles.travelConfirmBtn} onPress={confirmTravel}>
                <Text style={styles.travelConfirmText}>Travel</Text>
              </Pressable>
            </View>
          </View>
        </View>
      )}

      {/* Traveling loading overlay */}
      {isTraveling && (
        <View style={styles.travelOverlay}>
          <View style={styles.travelingBox}>
            <ActivityIndicator size="large" color="#e0d68a" />
            <Text style={styles.travelingText}>Traveling...</Text>
          </View>
        </View>
      )}

      {selectedPlayer && (
        <PlayerInteraction
          player={selectedPlayer}
          onClose={() => setSelectedPlayer(null)}
          onResult={() => {}}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: '#080c14',
    zIndex: 100,
    paddingTop: 50,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingHorizontal: 24,
    paddingBottom: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#1a1a2e',
  },
  title: {
    color: '#e0d68a',
    fontSize: 26,
    fontWeight: 'bold',
  },
  subtitle: {
    color: '#6ab04c',
    fontSize: 14,
    marginTop: 4,
  },
  closeButton: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 24,
    paddingVertical: 10,
    borderRadius: 20,
  },
  closeText: {
    color: '#e0d68a',
    fontWeight: 'bold',
    fontSize: 14,
  },
  scrollContent: {
    alignItems: 'center',
    paddingVertical: 20,
    paddingBottom: 60,
  },
  hexGrid: {
    position: 'relative',
    marginBottom: 12,
  },
  hexOuter: {
    position: 'absolute',
    justifyContent: 'center',
    alignItems: 'center',
  },
  hex: {
    width: '94%',
    height: '94%',
    justifyContent: 'center',
    alignItems: 'center',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#222',
    backgroundColor: '#111',
    // @ts-ignore - web only, pointy-top hexagon
    clipPath: 'polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)',
  },
  hexPlayer: {
    backgroundColor: '#1a3a1a',
    borderColor: '#6ab04c',
    borderWidth: 2,
  },
  hexKnown: {
    backgroundColor: '#2a4a2a',
    borderColor: '#4a7a4a',
  },
  hexHinted: {
    backgroundColor: '#1e2a3a',
    borderColor: '#2a3a4a',
  },
  hexDistant: {
    backgroundColor: '#0f1520',
    borderColor: '#1a2030',
  },
  hexFog: {
    backgroundColor: '#080c14',
    borderColor: '#121620',
  },
  hexSelected: {
    borderColor: '#e0d68a',
    borderWidth: 2,
  },
  hexIcon: {
    fontSize: 16,
  },
  hexIconFog: {
    fontSize: 12,
    opacity: 0.2,
  },
  hexName: {
    color: '#d4d4d4',
    fontSize: 7,
    textAlign: 'center',
    fontWeight: 'bold',
    marginTop: 0,
  },
  hexNameHinted: {
    color: '#556',
    fontWeight: 'normal',
    fontSize: 6,
  },
  hexYou: {
    fontSize: 8,
    position: 'absolute',
    bottom: 4,
  },
  hexLevel: {
    color: '#445',
    fontSize: 5,
    position: 'absolute',
    bottom: 2,
  },
  hexHint: {
    color: '#556',
    fontSize: 12,
    marginBottom: 16,
  },
  sectionTitle: {
    color: '#e0d68a',
    fontSize: 18,
    fontWeight: 'bold',
    alignSelf: 'flex-start',
    paddingHorizontal: 24,
    marginBottom: 10,
    marginTop: 8,
  },
  locationGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: 10,
    paddingHorizontal: 16,
    marginBottom: 16,
  },
  locationCard: {
    backgroundColor: '#1a1a2e',
    borderRadius: 12,
    padding: 12,
    alignItems: 'center',
    width: 110,
    borderWidth: 1,
    borderColor: '#2a2a3e',
  },
  locationDanger: {
    borderColor: '#c0392b',
    backgroundColor: '#1a1520',
  },
  disabled: {
    opacity: 0.4,
  },
  locationIcon: {
    fontSize: 26,
    marginBottom: 4,
  },
  locationName: {
    color: '#d4d4d4',
    fontSize: 11,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  locationDesc: {
    color: '#778',
    fontSize: 9,
    textAlign: 'center',
    marginTop: 2,
  },
  dangerText: {
    color: '#c0392b',
    fontWeight: 'bold',
  },
  npcRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: 8,
    paddingHorizontal: 16,
    marginBottom: 20,
  },
  npcCard: {
    backgroundColor: '#1a1a2e',
    borderRadius: 12,
    padding: 10,
    alignItems: 'center',
    width: 80,
    borderWidth: 1,
    borderColor: '#2a2a3e',
  },
  npcIcon: {
    fontSize: 24,
  },
  npcName: {
    color: '#d4d4d4',
    fontSize: 10,
    fontWeight: 'bold',
    textAlign: 'center',
    marginTop: 2,
  },
  npcRole: {
    color: '#667',
    fontSize: 8,
    textAlign: 'center',
  },
  legend: {
    paddingHorizontal: 24,
    marginTop: 8,
  },
  legendTitle: {
    color: '#556',
    fontSize: 12,
    fontWeight: 'bold',
    marginBottom: 6,
  },
  legendRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  legendDot: {
    width: 14,
    height: 14,
    borderRadius: 3,
    borderWidth: 1,
    borderColor: '#333',
  },
  legendText: {
    color: '#556',
    fontSize: 11,
    marginRight: 8,
  },
  playerCard: {
    backgroundColor: '#1a2030',
    borderRadius: 12,
    padding: 10,
    alignItems: 'center',
    width: 80,
    borderWidth: 1,
    borderColor: '#2a3a4a',
  },
  playerIcon: {
    fontSize: 24,
  },
  playerName: {
    color: '#7ecbf5',
    fontSize: 10,
    fontWeight: 'bold',
    textAlign: 'center',
    marginTop: 2,
  },
  playerKarma: {
    fontSize: 7,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  playerInfo: {
    color: '#556',
    fontSize: 8,
    textAlign: 'center',
  },
  playerHere: {
    color: '#27ae60',
    fontSize: 8,
    fontWeight: 'bold',
    marginTop: 2,
  },
  playerHint: {
    color: '#556',
    fontSize: 11,
    marginBottom: 8,
    paddingHorizontal: 24,
  },
  noPlayersText: {
    color: '#334',
    fontSize: 12,
    textAlign: 'center',
    marginVertical: 12,
  },
  travelOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.7)',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 200,
  },
  travelDialog: {
    backgroundColor: '#1a1a2e',
    borderRadius: 16,
    padding: 24,
    width: 280,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2a2a3e',
  },
  travelTitle: {
    color: '#e0d68a',
    fontSize: 18,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 12,
  },
  travelBiome: {
    color: '#d4d4d4',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 8,
  },
  travelTime: {
    color: '#778',
    fontSize: 13,
    textAlign: 'center',
    marginBottom: 20,
  },
  travelButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  travelCancel: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 24,
    paddingVertical: 10,
    borderRadius: 12,
  },
  travelCancelText: {
    color: '#aaa',
    fontWeight: 'bold',
    fontSize: 14,
  },
  travelConfirmBtn: {
    backgroundColor: '#27ae60',
    paddingHorizontal: 24,
    paddingVertical: 10,
    borderRadius: 12,
  },
  travelConfirmText: {
    color: '#fff',
    fontWeight: 'bold',
    fontSize: 14,
  },
  travelingBox: {
    backgroundColor: '#1a1a2e',
    borderRadius: 16,
    padding: 32,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2a2a3e',
  },
  travelingText: {
    color: '#e0d68a',
    fontSize: 18,
    fontWeight: 'bold',
    marginTop: 16,
  },
});
