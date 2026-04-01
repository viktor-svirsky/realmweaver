import React from 'react';
import { View, Text, Pressable, StyleSheet, ScrollView } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { sendAction } from '../api/socket';

interface Props {
  onCharacterSheet: () => void;
  onMap: () => void;
  onChat: () => void;
}

const iconMap: Record<string, string> = {
  sword: '\u2694\uFE0F',
  shield: '\u{1F6E1}\uFE0F',
  potion: '\u{1F9EA}',
  run: '\u{1F3C3}',
  magnifier: '\u{1F50D}',
  map: '\u{1F5FA}\uFE0F',
  bag: '\u{1F392}',
  campfire: '\u{1F525}',
  scroll: '\u{1F4DC}',
  coins: '\u{1FA99}',
  door: '\u{1F6AA}',
  tavern: '\u{1F37A}',
  anvil: '\u{2692}\uFE0F',
  market: '\u{1F3EA}',
  chapel: '\u26EA',
  dungeon: '\u{1F480}',
  npc: '\u{1F5E3}\uFE0F',
  ear: '\u{1F442}',
};

export default function ActionBar({ onCharacterSheet, onMap, onChat }: Props) {
  const quickActions = useGameStore((s) => s.quickActions);
  const isLoading = useGameStore((s) => s.isLoading);

  return (
    <View style={styles.container}>
      <View style={styles.topRow}>
        <Pressable style={styles.statsButton} onPress={onCharacterSheet}>
          <Text style={styles.statsText}>Stats</Text>
        </Pressable>
        <Pressable style={styles.mapButton} onPress={onMap}>
          <Text style={styles.statsText}>{'\u{1F5FA}\uFE0F'} Map</Text>
        </Pressable>
        <Pressable style={styles.mapButton} onPress={onChat}>
          <Text style={styles.statsText}>{'\u{1F4AC}'} Chat</Text>
        </Pressable>
      </View>
      <ScrollView style={styles.scrollArea} contentContainerStyle={styles.grid}>
        {quickActions.map((action) => (
          <Pressable
            key={action.id}
            style={[styles.actionButton, isLoading && styles.disabled]}
            onPress={() => {
              if (!isLoading) {
                useGameStore.getState().setLoading(true);
                sendAction(action.id, { target: action.id.replace('talk_', '') });
                // Reset loading after timeout (server response will also reset it)
                setTimeout(() => useGameStore.getState().setLoading(false), 15000);
              }
            }}
            disabled={isLoading}
          >
            <Text style={styles.actionIcon}>{iconMap[action.icon || ''] || '\u2728'}</Text>
            <Text style={styles.actionLabel}>{action.label}</Text>
          </Pressable>
        ))}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    maxHeight: 220,
    paddingHorizontal: 12,
    paddingBottom: 8,
  },
  topRow: {
    flexDirection: 'row',
    marginBottom: 8,
  },
  statsButton: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 20,
    paddingVertical: 8,
    borderRadius: 20,
  },
  statsText: {
    color: '#e0d68a',
    fontWeight: 'bold',
    fontSize: 14,
  },
  mapButton: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 20,
    paddingVertical: 8,
    borderRadius: 20,
    marginLeft: 8,
  },
  scrollArea: {
    flexGrow: 0,
  },
  grid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
  },
  actionButton: {
    backgroundColor: '#1a1a2e',
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderRadius: 12,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  disabled: {
    opacity: 0.4,
  },
  actionIcon: {
    fontSize: 16,
    marginRight: 6,
  },
  actionLabel: {
    color: '#d4d4d4',
    fontSize: 13,
  },
});
